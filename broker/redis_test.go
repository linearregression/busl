package broker

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	u "github.com/heroku/busl/util"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type RegistrarSuite struct {
	registrar Registrar
	uuid      u.UUID
}

type BrokerSuite struct {
	registrar Registrar
	uuid      u.UUID
	writer    io.WriteCloser
	reader    io.ReadCloser
}

var _ = Suite(&RegistrarSuite{})
var _ = Suite(&BrokerSuite{})

/*
 * Registrar Suite
 */
func (s *RegistrarSuite) SetUpTest(c *C) {
	s.registrar = NewRedisRegistrar()
	s.uuid, _ = u.NewUUID()
}

func (s *RegistrarSuite) TestRegisteredIsRegistered(c *C) {
	s.registrar.Register(s.uuid)
	c.Assert(s.registrar.IsRegistered(s.uuid), u.IsTrue)
}

func (s *RegistrarSuite) TestUnregisteredIsNotRegistered(c *C) {
	c.Assert(s.registrar.IsRegistered(s.uuid), u.IsFalse)
}

func (s *RegistrarSuite) TestUnregisteredErrNotRegistered(c *C) {
	_, err := NewReader(s.uuid)
	c.Assert(err, Equals, ErrNotRegistered)

	_, err = NewWriter(s.uuid)
	c.Assert(err, Equals, ErrNotRegistered)
}

func (s *RegistrarSuite) TestRegisteredNoError(c *C) {
	s.registrar.Register(s.uuid)
	_, err := NewReader(s.uuid)
	c.Assert(err, IsNil)

	_, err = NewWriter(s.uuid)
	c.Assert(err, IsNil)
}

/*
 * Broker Suite
 */

func (s *BrokerSuite) SetUpTest(c *C) {
	s.registrar = NewRedisRegistrar()
	s.uuid, _ = u.NewUUID()
	s.registrar.Register(s.uuid)
	s.writer, _ = NewWriter(s.uuid)
	s.reader, _ = NewReader(s.uuid)
}

func readstring(r io.Reader) string {
	buf, _ := ioutil.ReadAll(r)
	return string(buf)
}

func (s *BrokerSuite) TestRedisSubscribe(c *C) {
	done := make(chan bool)
	go func() {
		c.Assert(readstring(s.reader), Equals, "busl")
		done <- true
	}()

	s.writer.Write([]byte("busl"))
	s.writer.Close()
	<-done
}

func (s *BrokerSuite) TestRedisSubscribeReplay(c *C) {
	s.writer.Write([]byte("busl"))
	s.writer.Close()
	c.Assert(readstring(s.reader), Equals, "busl")
}

func (s *BrokerSuite) TestRedisSubscribeWithOffset(c *C) {
	s.writer.Write([]byte("busl"))
	s.writer.Close()

	s.reader.(io.Seeker).Seek(2, 0)
	defer s.reader.Close()
	c.Assert(readstring(s.reader), Equals, "sl")
}

func (s *BrokerSuite) TestRedisSubscribeOffsetLimits(c *C) {
	s.writer.Write([]byte("busl"))
	s.writer.Close()

	s.reader.(io.Seeker).Seek(4, 0)
	defer s.reader.Close()
	c.Assert(readstring(s.reader), Equals, "")

	s.reader.(io.Seeker).Seek(5, 0)
	io.Copy(os.Stdout, s.reader)
}

func (s *BrokerSuite) TestRedisSubscribeConcurrent(c *C) {
	pub := make(chan bool)
	done := make(chan bool)

	go func() {
		pub <- true
		s.writer.Write([]byte("busl"))
		s.writer.Close()
	}()

	go func() {
		<-pub
		c.Assert(readstring(s.reader), Equals, "busl")
		done <- true
	}()

	<-done
}

func (s *BrokerSuite) TestRedisReadFromClosed(c *C) {
	p := make([]byte, 10)

	// this read sets replayed = true
	s.reader.Read(p)
	s.writer.Close()

	// this read should short circuit with EOF
	_, err := s.reader.Read(p)
	c.Assert(err, Equals, io.EOF)
}
