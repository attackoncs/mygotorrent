package bencode

import(
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func objAssertStr(t *testing.T,expect string,o *BObject){
	assert.Equal(t,BSTR,o.type_)
	str,err:=o.Str()
	assert.Equal(t,nil,err)
	assert.Equal(t,expect,str)
}

func objAssertInt(t *testing.T,expect int,o *BObject){
	assert.Equal(t,BINT,o.type_)
	val,err:=o.Int()
	assert.Equal(t,nil,err)
	assert.Equal(t,expect,val)
}

func TestParseString(t *testing.T){
	var o *BObject
	in:="3:abc"
	buf:=bytes.NewBufferString(in)
	o,_=Parse(buf)
	objAssertStr(t,"abc",o)

	out:=bytes.NewBufferString("")
	assert.Equal(t,len(in),o.Bencode(out))
	assert.Equal(t,in,out.String())
}