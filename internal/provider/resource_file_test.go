package provider_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/lib/osrelease"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"path"
	"regexp"
	"strconv"
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
					ExpectError: regexp.MustCompile(`Error: file resource: file exists`),
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
