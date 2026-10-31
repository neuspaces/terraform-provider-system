package provider_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/lib/osrelease"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

var (
	testFileId uint32
)

type testFileConfig struct {
	fileName string
	userName string
}

func newTestFileConfig() testFileConfig {
	id := atomic.AddUint32(&testFileId, 1)

	return testFileConfig{
		userName: fmt.Sprintf("user%d", id),
		fileName: fmt.Sprintf("file-%d", id),
	}
}

func TestAccFile_create_content(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						// echo -n 'hello world!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
					),
				},
			},
		})
	})
}

func TestAccFile_create_content_sensitive(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content_sensitive", "hello s3cr3t!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						// echo -n 'hello s3cr3t!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "su6hNsFeImKJgWRpqtfzZw=="),
					),
				},
			},
		})
	})
}

func TestAccFile_create_source_file(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("source", "file://./test/hello-world.txt"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						// cat ./internal/provider/test/hello-world.txt | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
					),
				},
			},
		})
	})
}

func TestAccFile_create_source_file_without_schema(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("source", "./test/hello-world.txt"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						// cat ./internal/provider/test/hello-world.txt | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
					),
				},
			},
		})
	})
}

func TestAccFile_create_source_http(t *testing.T) {
	testConfig := newTestFileConfig()

	// TODO start an in-test http server which serves some static files instead of downloading from http://releases.hashicorp.com

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("source", "http://releases.hashicorp.com/terraform/1.6.3/terraform_1.6.3_SHA256SUMS"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						// curl -s 'https://releases.hashicorp.com/terraform/1.6.3/terraform_1.6.3_SHA256SUMS' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "5TkHdh91Xu5JW7tB6Id3NA=="),
					),
				},
			},
		})
	})
}

func TestAccFile_create_source_https(t *testing.T) {
	testConfig := newTestFileConfig()

	// TODO start an in-test http server which serves the files instead of downloading from https://releases.hashicorp.com

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("source", "https://releases.hashicorp.com/terraform/1.6.3/terraform_1.6.3_SHA256SUMS"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						// curl -s 'https://releases.hashicorp.com/terraform/1.6.3/terraform_1.6.3_SHA256SUMS' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "5TkHdh91Xu5JW7tB6Id3NA=="),
					),
				},
			},
		})
	})
}

func TestAccFile_update_mode(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "644"),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "666"),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "666"),
					),
				},
			},
		})
	})
}

func TestAccFile_update_uid(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		var updateUid int
		var updateUser string

		// Update to a pre-defined user of the target os
		switch target.Os.Id {
		case osrelease.AlpineId:
			updateUid = 2
			updateUser = "daemon"
		case osrelease.DebianId:
			updateUid = 1
			updateUser = "daemon"
		default:
			t.Skip()
		}

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							// root
							tfbuild.AttributeInt("uid", 0),
							tfbuild.AttributeString("content", "hello world!"),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							// Set to pre-defined user
							tfbuild.AttributeInt("uid", int64(updateUid)),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "uid", strconv.Itoa(updateUid)),
						resource.TestCheckResourceAttr("system_file.test", "user", updateUser),
					),
				},
			},
		})
	})
}

func TestAccFile_update_user(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		var updateUid int
		var updateUser string

		// Update to a pre-defined user of the target os
		switch target.Os.Id {
		case osrelease.AlpineId:
			updateUid = 2
			updateUser = "daemon"
		case osrelease.DebianId:
			updateUid = 1
			updateUser = "daemon"
		case osrelease.FedoraId:
			updateUid = 2
			updateUser = "daemon"
		default:
			t.Skip()
		}

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("user", "root"),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							// Set to pre-defined user
							tfbuild.AttributeString("user", updateUser),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "uid", strconv.Itoa(updateUid)),
						resource.TestCheckResourceAttr("system_file.test", "user", updateUser),
					),
				},
			},
		})
	})
}

func TestAccFile_update_gid(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						// TODO create own group
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							// tfbuild.AttributeInt("gid", 0),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeInt("gid", 4),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "gid", "4"),
						resource.TestCheckResourceAttr("system_file.test", "group", "adm"),
					),
				},
			},
		})
	})
}

func TestAccFile_update_group(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						// TODO create own group
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							// tfbuild.AttributeString("group", "root"),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("group", "adm"),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "gid", "4"),
						resource.TestCheckResourceAttr("system_file.test", "group", "adm"),
					),
				},
			},
		})
	})
}

func TestAccFile_update_content(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("content", "hello world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						// echo -n 'hello world!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("content", "hello universe!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						// echo -n 'hello universe!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "w0Y+MwVOASL+sUYDnI0Eww=="),
					),
				},
			},
		})
	})
}

func TestAccFile_update_content_sensitive(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("content_sensitive", "hello secure world!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						// echo -n 'hello secure world!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "TGLT7MnuVS/votmenbFc+w=="),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("content", "hello secure universe!"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						// echo -n 'hello secure universe!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "2Zf2/lbjchQgfqOpi07pbA=="),
					),
				},
			},
		})
	})
}

func TestAccFile_update_source(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("source", "./test/hello-world.txt"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						// cat ./internal/provider/test/hello-world.txt | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("source", "./test/hello-universe.txt"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						// cat ./internal/provider/test/hello-universe.txt | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "w0Y+MwVOASL+sUYDnI0Eww=="),
					),
				},
			},
		})
	})
}

func TestAccFile_fail_existing(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName)),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName)),
						testAccFileBlock("duplicate", testRunFilePath(target, testConfig.fileName)),
					)),
					ExpectError: regexp.MustCompile(`file resource\s+file exists`),
				},
			},
		})
	})
}

func TestAccFile_import(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("existing", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content", "first existing line\nsecond existing line"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.existing", "id", testRunFilePath(target, testConfig.fileName)),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("import", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content", "first existing line\nline in between\nsecond existing line"),
						),
					)),
					ImportState:        true,
					ImportStatePersist: false,
					ResourceName:       "system_file.import",
					ImportStateId:      testRunFilePath(target, testConfig.fileName),
					ImportStateVerify:  true,
					ImportStateVerifyIgnore: []string{
						"content",
					},
				},
			},
		})
	})
}

func TestAccFile_import_content(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("existing", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content", "first existing line\nsecond existing line"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.existing", "id", testRunFilePath(target, testConfig.fileName)),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("import", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content", "first existing line\nline in between\nsecond existing line"),
						),
					)),
					ImportState:        true,
					ImportStatePersist: false,
					ResourceName:       "system_file.import",
					ImportStateId:      fmt.Sprintf("%s:content", testRunFilePath(target, testConfig.fileName)),
					ImportStateVerify:  true,
					ImportStateVerifyIgnore: []string{
						"content",
					},
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) != 1 {
							return fmt.Errorf("expected 1 state")
						}

						importState := states[0]

						importStateContent := importState.Attributes["content"]
						expectedContent := "first existing line\nsecond existing line"
						if importStateContent != expectedContent {
							return fmt.Errorf("expected `system_file.import.content` attribute to be %s, got: %s", expectedContent, importStateContent)
						}

						return nil
					},
				},
			},
		})
	})
}

func TestAccFile_import_content_sensitive(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("existing", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content_sensitive", "first existing line\nsecond existing line"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.existing", "id", testRunFilePath(target, testConfig.fileName)),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("import", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content_sensitive", "first existing line\nline in between\nsecond existing line"),
						),
					)),
					ImportState:        true,
					ImportStatePersist: false,
					ResourceName:       "system_file.import",
					ImportStateId:      fmt.Sprintf("%s:content_sensitive", testRunFilePath(target, testConfig.fileName)),
					ImportStateVerify:  true,
					ImportStateVerifyIgnore: []string{
						"content_sensitive",
					},
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) != 1 {
							return fmt.Errorf("expected 1 state")
						}

						importState := states[0]

						importStateContentSensitive := importState.Attributes["content_sensitive"]
						expectedContentSensitive := "first existing line\nsecond existing line"
						if importStateContentSensitive != expectedContentSensitive {
							return fmt.Errorf("expected `system_file.import.content` attribute to be %s, got: %s", expectedContentSensitive, importStateContentSensitive)
						}

						return nil
					},
				},
			},
		})
	})
}

func testRunFilePath(target acctest.Target, p string) string {
	return path.Join(target.BasePath, p)
}

func testAccFileBlock(name string, path string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("path", path),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Resource("system_file", name, resourceAttrs...)
}

// testAccPreCreateFile creates a file on the target outside of Terraform, as a pre-existing object to adopt
func testAccPreCreateFile(t *testing.T, target acctest.Target, filePath string, content string) func() {
	return func() {
		err := client.NewFileClient(target.Provider.System).Create(context.Background(), client.File{
			Path: filePath, Uid: -1, Gid: -1, Content: strings.NewReader(content),
		})
		if err != nil {
			t.Fatalf("pre-create %s: %v", filePath, err)
		}
	}
}

// testAccCheckRemoteFileContent reads the file on the target over ssh and compares its content
func testAccCheckRemoteFileContent(target acctest.Target, filePath string, expected string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		f, err := client.NewFileClient(target.Provider.System, client.FileClientIncludeContent(true)).Get(context.Background(), filePath)
		if err != nil {
			return err
		}
		b, err := io.ReadAll(f.Content)
		if err != nil {
			return err
		}
		if string(b) != expected {
			return fmt.Errorf("remote file %s: expected content %q, got %q", filePath, expected, string(b))
		}
		return nil
	}
}

func TestAccFile_overwrite_existing(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		filePath := testRunFilePath(target, testConfig.fileName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					PreConfig: testAccPreCreateFile(t, target, filePath, "pre-existing content"),
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath,
							tfbuild.AttributeString("mode", "600"),
							tfbuild.AttributeString("content", "overwritten content!"),
							tfbuild.AttributeBool("overwrite", true),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", filePath),
						resource.TestCheckResourceAttr("system_file.test", "overwrite", "true"),
						resource.TestCheckResourceAttr("system_file.test", "mode", "600"),
						resource.TestCheckResourceAttr("system_file.test", "content", "overwritten content!"),
						testAccCheckRemoteFileContent(target, filePath, "overwritten content!"),
					),
				},
			},
		})
	})
}

func TestAccFile_overwrite_fail_type_mismatch(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		filePath := testRunFilePath(target, testConfig.fileName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						err := client.NewFolderClient(target.Provider.System).Create(context.Background(), client.Folder{Path: filePath, Uid: -1, Gid: -1})
						if err != nil {
							t.Fatalf("pre-create folder %s: %v", filePath, err)
						}
					},
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath,
							tfbuild.AttributeString("content", "must not be written"),
							tfbuild.AttributeBool("overwrite", true),
						),
					)),
					ExpectError: regexp.MustCompile(`file resource\s+file exists`),
				},
			},
		})
	})
}

func TestAccFile_overwrite_default_false_fails_on_existing(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		filePath := testRunFilePath(target, testConfig.fileName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					PreConfig: testAccPreCreateFile(t, target, filePath, "pre-existing content"),
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath, tfbuild.AttributeString("content", "x")),
					)),
					ExpectError: regexp.MustCompile(`file resource\s+file exists`),
				},
			},
		})
	})
}

// testAccFileRemoteSh runs a shell snippet on the target and returns trimmed stdout
func testAccFileRemoteSh(target acctest.Target, script string) (string, error) {
	res, err := client.ExecuteCommand(context.Background(), target.Provider.System, client.NewCommand(script))
	if err != nil {
		return "", err
	}
	if res.ExitCode != 0 {
		return "", fmt.Errorf("remote command failed (%d): %s", res.ExitCode, string(res.Stderr))
	}
	return strings.TrimSpace(string(res.Stdout)), nil
}

func TestAccFile_source_update_in_place(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		filePath := testRunFilePath(target, testConfig.fileName)
		linkPath := filePath + ".hardlink"
		localSource := filepath.Join(t.TempDir(), "source.txt")
		if err := os.WriteFile(localSource, []byte("content A\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		config := func() string {
			return tfbuild.FileString(tfbuild.File(
				acctest.ProviderConfigBlock(target.Configs.Default()),
				testAccFileBlock("test", filePath,
					tfbuild.AttributeString("mode", "644"),
					tfbuild.AttributeString("source", localSource),
				),
			))
		}

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config(),
					Check: resource.ComposeTestCheckFunc(
						// printf 'content A\n' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "1UlFuYpYWUrlqNhcarkN1Q=="),
						// Hard-link the remote file: an in-place rewrite keeps the inode, so the link follows the new content;
						// a delete+create would leave the link with the old content and a link count of 1.
						func(_ *terraform.State) error {
							_, err := testAccFileRemoteSh(target, fmt.Sprintf(`ln '%s' '%s' && stat -c %%h '%s'`, filePath, linkPath, filePath))
							return err
						},
					),
				},
				{
					PreConfig: func() {
						if err := os.WriteFile(localSource, []byte("content B\n"), 0o644); err != nil {
							t.Fatal(err)
						}
					},
					Config: config(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", filePath),
						// printf 'content B\n' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "W5nAopCKD+3fpMZCY7KW/Q=="),
						testAccCheckRemoteFileContent(target, filePath, "content B\n"),
						func(_ *terraform.State) error {
							out, err := testAccFileRemoteSh(target, fmt.Sprintf(`cat '%s'; stat -c %%h '%s'`, linkPath, filePath))
							if err != nil {
								return err
							}
							if out != "content B\n2" && out != "content B\n\n2" {
								return fmt.Errorf("expected in-place update (hard link shares content B, link count 2), got: %q", out)
							}
							return nil
						},
					),
				},
			},
		})
	})
}

func TestAccFile_update_content_to_source(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		filePath := testRunFilePath(target, testConfig.fileName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath, tfbuild.AttributeString("content", "hello world!")),
					)),
					Check: resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath, tfbuild.AttributeString("source", "./test/hello-universe.txt")),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "w0Y+MwVOASL+sUYDnI0Eww=="),
						resource.TestCheckResourceAttr("system_file.test", "content", ""),
					),
				},
			},
		})
	})
}

func TestAccFile_update_source_removed(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		filePath := testRunFilePath(target, testConfig.fileName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath, tfbuild.AttributeString("source", "./test/hello-world.txt")),
					)),
					Check: resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath, tfbuild.AttributeString("mode", "600")),
					)),
					Check: resource.ComposeTestCheckFunc(
						// the remote file keeps its content; only mode changes
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
						resource.TestCheckResourceAttr("system_file.test", "mode", "600"),
					),
				},
			},
		})
	})
}

func TestAccFile_overwrite_fail_symlink(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		filePath := testRunFilePath(target, testConfig.fileName)
		linkTargetPath := filePath + ".target"

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						// `path` is a symbolic link to a regular file elsewhere; `[ -e ]` and `[ -f ]` follow the link,
						// so without an explicit check for a link the content would be written to the link target
						if _, err := testAccFileRemoteSh(target, fmt.Sprintf(`printf 'link target content' > '%[1]s' && ln -s '%[1]s' '%[2]s'`, linkTargetPath, filePath)); err != nil {
							t.Fatalf("pre-create symlink %s: %v", filePath, err)
						}
					},
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath,
							tfbuild.AttributeString("mode", "600"),
							tfbuild.AttributeString("content", "must not be written"),
							tfbuild.AttributeBool("overwrite", true),
						),
					)),
					ExpectError: regexp.MustCompile(`file resource\s+file exists`),
				},
			},
		})

		// the link target must be untouched: nothing may be written through the symbolic link
		if err := testAccCheckRemoteFileContent(target, linkTargetPath, "link target content")(nil); err != nil {
			t.Error(err)
		}
	})
}

func TestAccFile_update_content_to_content_sensitive(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		filePath := testRunFilePath(target, testConfig.fileName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath, tfbuild.AttributeString("content", "hello world!")),
					)),
					// echo -n 'hello world!' | openssl dgst -binary -md5 | openssl base64
					Check: resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", filePath, tfbuild.AttributeString("content_sensitive", "hello s3cr3t!")),
					)),
					Check: resource.ComposeTestCheckFunc(
						// echo -n 'hello s3cr3t!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "su6hNsFeImKJgWRpqtfzZw=="),
						resource.TestCheckResourceAttr("system_file.test", "content", ""),
						testAccCheckRemoteFileContent(target, filePath, "hello s3cr3t!"),
					),
				},
			},
		})
	})
}
