package libvirt

import libvirt "github.com/libvirt/libvirt-go"

// StreamIO libvirt struct
type StreamIO struct {
	Stream libvirt.Stream
}

// NewStreamIO returns libvirt StreamIO
func NewStreamIO(s libvirt.Stream) *StreamIO {
	return &StreamIO{Stream: s}
}

func (sio *StreamIO) Read(p []byte) (int, error) {
	return sio.Stream.Recv(p)
}

func (sio *StreamIO) Write(p []byte) (int, error) {
	return sio.Stream.Send(p)
}
