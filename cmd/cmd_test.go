package cmd

import "testing"

func TestInstall(t *testing.T) {
	default_conf_file_path = "/tmp/test.json"
	err := install_service()
	if err != nil {
		t.Fatal(err)
	}
}
