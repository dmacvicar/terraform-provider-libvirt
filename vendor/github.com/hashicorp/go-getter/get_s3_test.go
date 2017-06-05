package getter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

func init() {
	// These are well known restricted IAM keys to a HashiCorp-managed bucket
	// in a private AWS account that only has access to the open source test
	// resources.
	//
	// We do the string concat below to avoid AWS autodetection of a key. This
	// key is locked down an IAM policy that is read-only so we're purposely
	// exposing it.
	os.Setenv("AWS_ACCESS_KEY", "AKIAJCTNQ"+"IOBWAYXKGZA")
	os.Setenv("AWS_SECRET_KEY", "jcQOTYdXNzU5MO"+"5ExqbE1U995dIfKCKQtiVobMvr")
}

func TestS3Getter_impl(t *testing.T) {
	var _ Getter = new(S3Getter)
}

func TestS3Getter(t *testing.T) {
	g := new(S3Getter)
	dst := tempDir(t)

	// With a dir that doesn't exist
	err := g.Get(
		dst, testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/folder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestS3Getter_subdir(t *testing.T) {
	g := new(S3Getter)
	dst := tempDir(t)

	// With a dir that doesn't exist
	err := g.Get(
		dst, testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/folder/subfolder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	subPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(subPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestS3Getter_GetFile(t *testing.T) {
	g := new(S3Getter)
	dst := tempFile(t)

	// Download
	err := g.GetFile(
		dst, testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/folder/main.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "# Main\n")
}

func TestS3Getter_GetFile_badParams(t *testing.T) {
	g := new(S3Getter)
	dst := tempFile(t)

	// Download
	err := g.GetFile(
		dst,
		testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/folder/main.tf?aws_access_key_id=foo&aws_access_key_secret=bar&aws_access_token=baz"))
	if err == nil {
		t.Fatalf("expected error, got none")
	}

	if reqerr, ok := err.(awserr.RequestFailure); !ok || reqerr.StatusCode() != 403 {
		t.Fatalf("expected InvalidAccessKeyId error")
	}
}

func TestS3Getter_GetFile_notfound(t *testing.T) {
	g := new(S3Getter)
	dst := tempFile(t)

	// Download
	err := g.GetFile(
		dst, testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/folder/404.tf"))
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}

func TestS3Getter_ClientMode_dir(t *testing.T) {
	g := new(S3Getter)

	// Check client mode on a key prefix with only a single key.
	mode, err := g.ClientMode(
		testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/folder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeDir {
		t.Fatal("expect ClientModeDir")
	}
}

func TestS3Getter_ClientMode_file(t *testing.T) {
	g := new(S3Getter)

	// Check client mode on a key prefix which contains sub-keys.
	mode, err := g.ClientMode(
		testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/folder/main.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestS3Getter_ClientMode_notfound(t *testing.T) {
	g := new(S3Getter)

	// Check the client mode when a non-existent key is looked up. This does not
	// return an error, but rather should just return the file mode so that S3
	// can return an appropriate error later on. This also checks that the
	// prefix is handled properly (e.g., "/fold" and "/folder" don't put the
	// client mode into "dir".
	mode, err := g.ClientMode(
		testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/fold"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestS3Getter_ClientMode_collision(t *testing.T) {
	g := new(S3Getter)

	// Check that the client mode is "file" if there is both an object and a
	// folder with a common prefix (i.e., a "collision" in the namespace).
	mode, err := g.ClientMode(
		testURL("https://s3.amazonaws.com/hc-oss-test/go-getter/collision/foo"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}
