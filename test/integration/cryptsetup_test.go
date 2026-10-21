//go:build integration

package integration

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/hetznercloud/csi-driver/internal/volumes"
)

func TestCryptSetup(t *testing.T) {
	if !runTestInDockerImage(t, true) {
		return
	}

	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cryptSetup := volumes.NewCryptSetup(logger)
	name := "fake"
	device, _, err := createFakeDevice(name, 32)
	if err != nil {
		t.Fatal(err)
	}
	passphrase := "passphrase"

	if err := cryptSetup.Format(ctx, device, passphrase); err != nil {
		t.Fatal(err)
	}
	decryptedName := name + "-decrypted"

	if err := cryptSetup.Open(ctx, device, decryptedName, passphrase, volumes.OpenOpts{DisableWorkqueues: true}); err != nil {
		t.Fatal(err)
	}
	decryptedDevice := "/dev/mapper/" + decryptedName
	defer runCmd("cryptsetup", "luksClose", decryptedName)

	status, err := runCmd("cryptsetup", "status", decryptedName)
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"discards", "no_read_workqueue", "no_write_workqueue"} {
		if !strings.Contains(status, flag) {
			t.Fatalf("expected dm-crypt flag %s in status: %s", flag, status)
		}
	}

	if _, err := runCmd("mkfs.ext4", decryptedDevice); err != nil {
		t.Fatal(err)
	}
	decryptedMount := "/mnt/" + name
	if err := os.MkdirAll(decryptedMount, 0o775); err != nil {
		t.Fatal(err)
	}
	if _, err := runCmd("mount", "-t", "ext4", decryptedDevice, decryptedMount); err != nil {
		t.Fatal(err)
	}
	defer runCmd("umount", decryptedMount)

	if _, err := runCmd("umount", decryptedMount); err != nil {
		t.Fatal(err)
	}
	if err := cryptSetup.Close(ctx, decryptedName); err != nil {
		t.Fatal(err)
	}
}
