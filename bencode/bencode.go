package bencode

import (
	"bufio"
	"errors"
	"io"
)

var(
	ErrType=errors.New("type error")
	ErrNum=errors.New("expect number")
	ErrCol=errors.New("expect colon")
	ErrEpI=errors.New("expect char i")
	ErrEpE=errors.New("expect char e")
	ErrIvd=errors.New("invalid bencode")
)

type BType uint8

//BenCode4种编码类型
const (
	BSTR  BType = 0x01
	BINT  BType = 0x02
	BLIST BType = 0x03
	BDICT BType = 0x04
)

//不同类型对应不同value，go已经支持泛型，等价void*
type BValue interface{}

type BObject struct {
	type_ BType
	val_  BValue
}

func (o *BObject) Str() (string, error) {
	if o.type_ != BSTR {
		return "", ErrType
	}
	return o.val_.(string), nil
}

func (o *BObject) Int() (int, error) {
	if o.type_ != BINT {
		return 0, ErrType
	}
	return o.val_.(int), nil
}

func (o *BObject) List() ([]*BObject, error) {
	if o.type_ != BLIST {
		return nil, ErrType
	}
	return o.val_.([]*BObject), nil //指帧的slice，避免大量拷贝
}

func (o *BObject) Dict() (map[string]*BObject, error) {
	if o.type_ != BDICT {
		return nil, ErrType
	}
	return o.val_.(map[string]*BObject), nil
}

//BObject转Bencode返回长度
func (o *BObject) Bencode(w io.Writer) int {
	bw,ok:=w.(*bufio.Writer)
	if !ok{
		bw=bufio.NewWriter(w)
	}
	wLen:=0
	switch o.type_ {
	case BSTR:
		str,_:=o.Str()
		wLen+=EncodeString(bw,str)
	case BINT:
		val,_:=o.Int()
		wLen+=EncodeInt(bw,val)
	case BLIST:
		bw.WriteByte('l')
		list,_:=o.List()
		for _,elem:=range list{
			wLen+=elem.Bencode(bw)//递归调用
		}
		bw.WriteByte('e')
		wLen+=2
	case BDICT:
		bw.WriteByte('d')
		dict,_:=o.Dict()
		for k,v:=range dict{
			wLen+=EncodeString(bw,k)
			wLen+=v.Bencode(bw)//递归调用
		}
		bw.WriteByte('e')
		wLen+=2
	}
	bw.Flush()
	return wLen
}

func checkNum(data byte) bool {
	return data>='0'&&data<='9'
}

func readDecimal(r *bufio.Reader) (val int, len int) {
	sign :=1
	b,_:=r.ReadByte()
	len++
	if b=='-'{
		sign=-1
		b,_=r.ReadByte()
		len++
	}
	for{
		if !checkNum(b){
			r.UnreadByte()
			len--
			return sign*val,len
		}
		val=val*10+int(b-'0')
		b,_=r.ReadByte()
		len++
	}
}

//写数字，先特判0和负数，再先乘找到最解决val的dividend数再除，通过处数和商数求得字符串
func writeDecimal(w *bufio.Writer, val int) (len int) {
	if val==0{
		w.WriteByte('0')
		len++
		return
	}
	if val<0{
		w.WriteByte('-')
		len++
		val*=-1
	}
	dividend:=1
	for{
		if dividend>val{
			dividend/=10
			break
		}
		dividend*=10
	}

	for{
		num:=byte(val/dividend)
		w.WriteByte('0'+num)
		len++
		if dividend==1{
			return
		}
		val%=dividend
		dividend/=10
	}
}

//3:abc
func EncodeString(w io.Writer, val string) int {
	strlen:=len(val)
	bw:=bufio.NewWriter(w)
	wLen:=writeDecimal(bw,strlen)//写入长度的字符表示
	bw.WriteByte(':')//写入冒号
	wLen++
	bw.WriteString(val)//写入字符串
	wLen+=strlen
	err:=bw.Flush()
	if err!=nil{
		return 0
	}
	return wLen
}

func DecodeString(r io.Reader) (val string,err error) {
	br,ok:=r.(*bufio.Reader)
	if !ok{
		br=bufio.NewReader(r)
	}
	num,len:=readDecimal(br)
	if len==0{
		return val,ErrNum
	}
	b,err:=br.ReadByte()
	if b!=':'{
		return val,ErrCol
	}
	buf:=make([]byte,num)
	_,err=io.ReadAtLeast(br,buf,num)
	val=string(buf)
	return
}

//i-199e，在数字val前加上字符'i'表示整型数，最后加上字符'e'作为后缀，返回长度
func EncodeInt(w io.Writer, val int) int {
	wLen:=0
	bw:=bufio.NewWriter(w)
	bw.WriteByte('i')
	wLen++
	nLen:=writeDecimal(bw,val)
	wLen+=nLen
	bw.WriteByte('e')
	wLen++

	err:=bw.Flush()
	if err!=nil{
		return 0
	}
	return wLen
}

//EncodeInt的逆过程，返回解密的值
func DecodeInt(r io.Reader) (val int,err error) {
	br,ok:=r.(*bufio.Reader)
	if !ok{
		br=bufio.NewReader(r)
	}
	b,err:=br.ReadByte()
	if b!='i'{
		return val,ErrEpI
	}
	val,_=readDecimal(br)
	b,err=br.ReadByte()
	if b!='e'{
		return val,ErrEpE
	}
	return
}
