healthcheck {
    address = ":8080"
    path    = "/healthz"
}

userprovider ldap {
    address = "glauth:8080"
}

tcpproxy ":80" {
    destination = "127.0.0.1:3000"
}

tcpproxy ":443" {
    destination = "127.0.0.1:3000"
}

httproxy ":8000" {
    downstream filebrowser {
        upstream = "filebrowser"

        rule path-prefix {
            path = "/storage"
        }
        authorizer cookie {
            key = "access"
        }
    }

    upstream filebrowser {
        address = "filebrowser:80"

        authorizer header {
            userID   = "X-User-ID"
            username = "X-Username"
        }
    }
}