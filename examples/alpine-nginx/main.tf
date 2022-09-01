terraform {
  required_providers {
    system = {
      source = "registry.terraform.io/neuspaces/system"
    }
  }
}

provider "system" {
  ssh {
    host        = "example.host.local"
    port        = 22
    user        = "root"
    private_key = file("./root.key")
  }
}

# Variables

variable "virtual_host_name" {
  default = "example.com"
}

# Install nginx

resource "system_packages_apk" "packages" {
  package {
    name = "nginx"
  }
}

# Create virtual host document root

resource "system_folder" "document_root" {
  path = "/var/www/${var.virtual_host_name}"
}

resource "system_folder" "document_root_html" {
  path = "${system_folder.document_root.path}/html"
}

# Create index page

resource "system_file" "index_html" {
  path = "${system_folder.document_root_html.path}/index.html"
  content = trimspace(<<EOT
<html>
    <head>
        <title>Welcome to ${var.virtual_host_name}!</title>
    </head>
    <body>
        <h1>You have successfully configured nginx using the terraform provider system!</h1>
    </body>
</html>
EOT
  )
}

# Create virtual host configuration

resource "system_file" "virtual_host" {
  path = "/etc/nginx/http.d/${var.virtual_host_name}.conf"
  content = trimspace(<<EOT
server {
        listen 80;
        listen [::]:80;

        server_name ${var.virtual_host_name};

        root /var/www/${var.virtual_host_name}/html;
        index index.html;

        location / {
                try_files $uri $uri/ =404;
        }
}
EOT
  )
}

# Enable and start nginx service

resource "system_service_openrc" "nginx" {
  name = "nginx"

  enabled = true
  status  = "started"

  depends_on = [
    system_packages_apk.packages,
  ]
}
