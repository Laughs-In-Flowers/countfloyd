package feature

import (
	cr "crypto/rand"

	"github.com/Laughs-In-Flowers/xrr"
)

var (
	ExistsError       = xrr.Xrror("A %s named %s already exists.").Out
	DoesNotExistError = xrr.Xrror("A %s named %s does not exist.").Out
	NotFoundError     = xrr.Xrror("%s named %s not found").Out
)

type uuid [16]byte

var halfbyte2hexchar = []byte{
	48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 97, 98, 99, 100, 101, 102,
}

func (u uuid) String() string {
	b := [36]byte{}

	for i, n := range []int{
		0, 2, 4, 6,
		9, 11,
		14, 16,
		19, 21,
		24, 26, 28, 30, 32, 34,
	} {
		b[n] = halfbyte2hexchar[(u[i]>>4)&0x0f]
		b[n+1] = halfbyte2hexchar[u[i]&0x0f]
	}

	b[8] = '-'
	b[13] = '-'
	b[18] = '-'
	b[23] = '-'

	return string(b[:])
}

func v4() (uuid, error) {
	u := uuid{}

	_, err := cr.Read(u[:])
	if err != nil {
		return u, err
	}

	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F

	return u, nil
}

func genUUID() string {
	u, err := v4()
	if err != nil {
		return err.Error()
	}
	return u.String()
}
