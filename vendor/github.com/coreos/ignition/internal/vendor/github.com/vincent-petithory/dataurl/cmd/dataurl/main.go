package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path"

	"github.com/vincent-petithory/dataurl"
)

var (
	performDecode bool
	asciiEncoding bool
	mimetype      string
)

func init() {
	const decodeUsage = "decode data instead of encoding"
	flag.BoolVar(&performDecode, "decode", false, decodeUsage)
	flag.BoolVar(&performDecode, "d", false, decodeUsage)

	const mimetypeUsage = "force the mimetype of the data to encode to this value"
	flag.StringVar(&mimetype, "mimetype", "", mimetypeUsage)
	flag.StringVar(&mimetype, "m", "", mimetypeUsage)

	const asciiUsage = "encode data using ascii instead of base64"
	flag.BoolVar(&asciiEncoding, "ascii", false, asciiUsage)
	flag.BoolVar(&asciiEncoding, "a", false, asciiUsage)

	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			`dataurl - Encode or decode dataurl data and print to standard output

Usage: dataurl [OPTION]... [FILE]

  dataurl encodes or decodes FILE or standard input if FILE is - or omitted, and prints to standard output.
  Unless -mimetype is used, when FILE is specified, dataurl will attempt to detect its mimetype using Go's mime.TypeByExtension (http://golang.org/pkg/mime/#TypeByExtension). If this fails or data is read from STDIN, the mimetype will default to application/octet-stream.

Options:
`)
		flag.PrintDefaults()
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	var (
		in               io.Reader
		out              = os.Stdout
		encoding         = dataurl.EncodingBase64
		detectedMimetype string
	)
	switch n := flag.NArg(); n {
	case 0:
		in = os.Stdin
	case 1:
		if flag.Arg(0) == "-" {
			in = os.Stdin
			return
		}
		if f, err := os.Open(flag.Arg(0)); err != nil {
			log.Fatal(err)
		} else {
			in = f
			defer f.Close()
		}
		ext := path.Ext(flag.Arg(0))
		detectedMimetype = mime.TypeByExtension(ext)
	}

	switch {
	case mimetype == "" && detectedMimetype == "":
		mimetype = "application/octet-stream"
	case mimetype == "" && detectedMimetype != "":
		mimetype = detectedMimetype
	}

	if performDecode {
		if err := decode(in, out); err != nil {
			log.Fatal(err)
		}
	} else {
		if asciiEncoding {
			encoding = dataurl.EncodingASCII
		}
		if err := encode(in, out, encoding, mimetype); err != nil {
			log.Fatal(err)
		}
	}
}

func decode(in io.Reader, out io.Writer) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	du, err := dataurl.Decode(in)
	if err != nil {
		return
	}

	_, err = out.Write(du.Data)
	if err != nil {
		return
	}
	return
}

func encode(in io.Reader, out io.Writer, encoding string, mediatype string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = fmt.Errorf("%v", e)
			}
			return
		}
	}()
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return
	}

	du := dataurl.New(b, mediatype)
	du.Encoding = encoding

	_, err = du.WriteTo(out)
	if err != nil {
		return
	}
	return
}
