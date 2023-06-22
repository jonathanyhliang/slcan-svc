package slcansvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabase(t *testing.T) {
	var m Message

	// call GetData(), no data found
	m, err := db.GetData(0x7ff)
	assert.Empty(t, m)
	assert.NotEqual(t, nil, err)

	// call PostData(), write id:0x7ff succeed
	m = Message{0x7ff, "200rpm"}
	err = db.PostData(m)
	assert.Equal(t, nil, err)

	// call GetData(), read id:0x7ff succeed
	m, err = db.GetData(0x7ff)
	assert.Equal(t, m, Message{0x7ff, "200rpm"})
	assert.Equal(t, nil, err)

	// call PostData(), data already exists
	m = Message{0x7ff, "200rpm"}
	err = db.PostData(m)
	assert.NotEqual(t, nil, err)

	// call PutData(), write id:0x7ff succeed
	m = Message{0x7ff, "201rpm"}
	err = db.PutData(0x7ff, m)
	assert.Equal(t, nil, err)

	// call GetData(), read id:0x7ff succeed
	m, err = db.GetData(0x7ff)
	assert.Equal(t, m, Message{0x7ff, "201rpm"})
	assert.Equal(t, nil, err)

	// call DeleteData(), delete id:0x7ff succeed
	err = db.DeleteData(0x7ff)
	assert.Equal(t, nil, err)

	// call GetData(), no data found
	m, err = db.GetData(0x7ff)
	assert.Empty(t, m)
	assert.NotEqual(t, nil, err)

	// call PutData(), no data found
	m = Message{0x7ff, "201rpm"}
	err = db.PutData(0x7ff, m)
	assert.NotEqual(t, nil, err)

	// call DeleteData(), no data found
	err = db.DeleteData(0x7ff)
	assert.NotEqual(t, nil, err)
}
