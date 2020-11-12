package libvirt

import libvirtc "github.com/libvirt/libvirt-go"

// StreamIO libvirt struct
type StreamIO struct {
	Stream libvirtc.Stream
}

// NewStreamIO returns libvirt StreamIO
func NewStreamIO(s libvirtc.Stream) *StreamIO {
	return &StreamIO{Stream: s}
}

func (sio *StreamIO) Read(p []byte) (int, error) {
	return sio.Stream.Recv(p)
}

func (sio *StreamIO) Write(p []byte) (int, error) {
	return sio.Stream.Send(p)
}
