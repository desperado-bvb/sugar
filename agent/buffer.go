package agent

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

var (
	bufcnt int64
)

const (
	defaultBufferSize     = 1024 * 16
	defaultReadBlockSize  = 8192
	defaultWriteBlockSize = 8192
)

type sequence struct {
	cursor,

	gate,

	p2, p3, p4, p5, p6, p7 int64
}

func newSequence() *sequence {
	return &sequence{}
}

func (this *sequence) get() int64 {
	return atomic.LoadInt64(&this.cursor)
}

func (this *sequence) set(seq int64) {
	atomic.StoreInt64(&this.cursor, seq)
}

type buffer struct {
	id int64

	buf []byte
	tmp []byte

	size int64
	mask int64

	done int64

	pseq *sequence
	cseq *sequence

	pcond *sync.Cond
	ccond *sync.Cond

	cwait int64
	pwait int64
}

func newBuffer(size int64) (*buffer, error) {
	if size < 0 {
		return nil, bufio.ErrNegativeCount
	}

	if size == 0 {
		size = defaultBufferSize
	}

	if !powerOfTwo64(size) {
		return nil, fmt.Errorf("Size must be power of two. Try %d.", roundUpPowerOfTwo64(size))
	}

	if size < 2*defaultReadBlockSize {
		return nil, fmt.Errorf("Size must at least be %d. Try %d.", 2*defaultReadBlockSize, 2*defaultReadBlockSize)
	}

	return &buffer{
		id:    atomic.AddInt64(&bufcnt, 1),
		buf:   make([]byte, size),
		size:  size,
		mask:  size - 1,
		pseq:  newSequence(),
		cseq:  newSequence(),
		pcond: sync.NewCond(new(sync.Mutex)),
		ccond: sync.NewCond(new(sync.Mutex)),
		cwait: 0,
		pwait: 0,
	}, nil
}

func (this *buffer) ID() int64 {
	return this.id
}

func (this *buffer) Close() error {
	atomic.StoreInt64(&this.done, 1)

	this.pcond.L.Lock()
	this.pcond.Broadcast()
	this.pcond.L.Unlock()

	this.pcond.L.Lock()
	this.ccond.Broadcast()
	this.pcond.L.Unlock()

	return nil
}

func (this *buffer) Len() int {
	cpos := this.cseq.get()
	ppos := this.pseq.get()
	return int(ppos - cpos)
}

func (this *buffer) ReadFrom(r io.Reader) (int64, error) {
	defer this.Close()

	total := int64(0)

	for {
		if this.isDone() {
			return total, io.EOF
		}

		start, cnt, err := this.waitForWriteSpace(defaultReadBlockSize)
		if err != nil {
			return 0, err
		}

		pstart := start & this.mask
		pend := pstart + int64(cnt)
		if pend > this.size {
			pend = this.size
		}

		//conn.SetReadDeadline(time.Now().Add(time.Second * keepAliveTime))		
		n, err := r.Read(this.buf[pstart:pend])
		if n > 0 {
			total += int64(n)
			_, err := this.WriteCommit(n)
			if err != nil {
				return total, err
			}
		}

		if err != nil {
			if this.isDone() {
				return total, io.EOF
			}

			return total, err
		}
	}
}

func (this *buffer) WriteTo(w io.Writer) (int64, error) {
	defer this.Close()

	total := int64(0)

	for {
		if this.isDone() {
			return total, io.EOF
		}

		p, err := this.ReadPeek(defaultWriteBlockSize)

		if len(p) > 0 {
			n, err := w.Write(p)
			total += int64(n)
			//glog.Debugf("Wrote %d bytes, totaling %d bytes", n, total)

			if err != nil {
				return total, err
			}

			_, err = this.ReadCommit(n)
			if err != nil {
				return total, err
			}
		}

		if err != ErrBufferInsufficientData && err != nil {
			return total, err
		}
	}
}

func (this *buffer) Read(p []byte) (int, error) {
	if this.isDone() && this.Len() == 0 {
		return 0, io.EOF
	}

	pl := int64(len(p))

	for {
		cpos := this.cseq.get()
		ppos := this.pseq.get()
		cindex := cpos & this.mask

		if cpos+pl < ppos {
			n := copy(p, this.buf[cindex:])

			this.cseq.set(cpos + int64(n))
			this.pcond.L.Lock()
			this.pcond.Broadcast()
			this.pcond.L.Unlock()

			return n, nil
		}

		if cpos < ppos {
			b := ppos - cpos

			var n int

			if cindex+b < this.size {
				n = copy(p, this.buf[cindex:cindex+b])
			} else {
				n = copy(p, this.buf[cindex:])
			}

			this.cseq.set(cpos + int64(n))
			this.pcond.L.Lock()
			this.pcond.Broadcast()
			this.pcond.L.Unlock()
			return n, nil
		}

		this.ccond.L.Lock()
		for ppos = this.pseq.get(); cpos >= ppos; ppos = this.pseq.get() {
			if this.isDone() {
				return 0, io.EOF
			}

			this.cwait++
			this.ccond.Wait()
		}
		this.ccond.L.Unlock()
	}
}

func (this *buffer) Write(p []byte) (int, error) {
	if this.isDone() {
		return 0, io.EOF
	}

	start, _, err := this.waitForWriteSpace(len(p))
	if err != nil {
		return 0, err
	}

	total := ringCopy(this.buf, p, int64(start)&this.mask)

	this.pseq.set(start + int64(len(p)))
	this.ccond.L.Lock()
	this.ccond.Broadcast()
	this.ccond.L.Unlock()

	return total, nil
}

func (this *buffer) ReadPeek(n int) ([]byte, error) {
	if int64(n) > this.size {
		return nil, bufio.ErrBufferFull
	}

	if n < 0 {
		return nil, bufio.ErrNegativeCount
	}

	cpos := this.cseq.get()
	ppos := this.pseq.get()

	this.ccond.L.Lock()
	for ; cpos >= ppos; ppos = this.pseq.get() {
		if this.isDone() {
			return nil, io.EOF
		}

		this.cwait++
		this.ccond.Wait()
	}
	this.ccond.L.Unlock()

	m := ppos - cpos
	err := error(nil)

	if m >= int64(n) {
		m = int64(n)
	} else {
		err = ErrBufferInsufficientData
	}

	if cpos+m <= ppos {
		cindex := cpos & this.mask
		if cindex+m > this.size {
			this.tmp = this.tmp[0:0]

			l := len(this.buf[cindex:])
			this.tmp = append(this.tmp, this.buf[cindex:]...)
			this.tmp = append(this.tmp, this.buf[0:m-int64(l)]...)
			return this.tmp, err
		} else {
			return this.buf[cindex : cindex+m], err
		}
	}

	return nil, ErrBufferInsufficientData
}

func (this *buffer) ReadWait(n int) ([]byte, error) {
	if int64(n) > this.size {
		return nil, bufio.ErrBufferFull
	}

	if n < 0 {
		return nil, bufio.ErrNegativeCount
	}

	cpos := this.cseq.get()
	ppos := this.pseq.get()

	next := cpos + int64(n)

	this.ccond.L.Lock()
	for ; next > ppos; ppos = this.pseq.get() {
		if this.isDone() {
			return nil, io.EOF
		}

		this.ccond.Wait()
	}
	this.ccond.L.Unlock()

	cindex := cpos & this.mask

	if cindex+int64(n) > this.size {
		this.tmp = this.tmp[0:0]

		l := len(this.buf[cindex:])
		this.tmp = append(this.tmp, this.buf[cindex:]...)
		this.tmp = append(this.tmp, this.buf[0:n-l]...)
		return this.tmp[:n], nil
	}

	return this.buf[cindex : cindex+int64(n)], nil
}

func (this *buffer) ReadCommit(n int) (int, error) {
	if int64(n) > this.size {
		return 0, bufio.ErrBufferFull
	}

	if n < 0 {
		return 0, bufio.ErrNegativeCount
	}

	cpos := this.cseq.get()
	ppos := this.pseq.get()

	if cpos+int64(n) <= ppos {
		this.cseq.set(cpos + int64(n))
		this.pcond.L.Lock()
		this.pcond.Broadcast()
		this.pcond.L.Unlock()
		return n, nil
	}

	return 0, ErrBufferInsufficientData
}

func (this *buffer) WriteWait(n int) ([]byte, bool, error) {
	start, cnt, err := this.waitForWriteSpace(n)
	if err != nil {
		return nil, false, err
	}

	pstart := start & this.mask
	if pstart+int64(cnt) > this.size {
		return this.buf[pstart:], true, nil
	}

	return this.buf[pstart : pstart+int64(cnt)], false, nil
}

func (this *buffer) WriteCommit(n int) (int, error) {
	start, cnt, err := this.waitForWriteSpace(n)
	if err != nil {
		return 0, err
	}

	this.pseq.set(start + int64(cnt))

	this.ccond.L.Lock()
	this.ccond.Broadcast()
	this.ccond.L.Unlock()

	return cnt, nil
}

func (this *buffer) waitForWriteSpace(n int) (int64, int, error) {
	if this.isDone() {
		return 0, 0, io.EOF
	}

	ppos := this.pseq.get()

	next := ppos + int64(n)

	gate := this.pseq.gate

	wrap := next - this.size

	if wrap > gate || gate > ppos {
		var cpos int64
		this.pcond.L.Lock()
		for cpos = this.cseq.get(); wrap > cpos; cpos = this.cseq.get() {
			if this.isDone() {
				return 0, 0, io.EOF
			}

			this.pwait++
			this.pcond.Wait()
		}

		this.pseq.gate = cpos
		this.pcond.L.Unlock()
	}

	return ppos, n, nil
}

func (this *buffer) isDone() bool {
	if atomic.LoadInt64(&this.done) == 1 {
		return true
	}

	return false
}

func ringCopy(dst, src []byte, start int64) int {
	n := len(src)

	i, l := 0, 0

	for n > 0 {
		l = copy(dst[start:], src[i:])
		i += l
		n -= l

		if n > 0 {
			start = 0
		}
	}

	return i
}

func powerOfTwo64(n int64) bool {
	return n != 0 && (n&(n-1)) == 0
}

func roundUpPowerOfTwo64(n int64) int64 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++

	return n
}
