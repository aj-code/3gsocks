# 3gsocks

This is a multi-platform reverse connect remote socks5 network pivot.

* Socks5 based network pivoting
* TLS transport with cert pinning
* Should run on damn near anything (probably ymmv)
* Precompiled binaries available (under dist/) if you're that way inclined

Inspired by some socks functionality I found that one time in some malware I was reverse.


## Usage

Spin up the server on your local machine (or your c2 server, whatever) which will listen on two ports. One port accepts TLS connections from the remote client and the other is a socks5 server which you can pipe whatever you like through.

If you're on some sort of linux you can do the following:

Generate some self signed certs

`$ openssl req -x509 -nodes -newkey rsa:4096 -keyout key.pem -out cert.pem -days 90`

Run the server (on linux)

`$ ./3gsocks_server_linux_amd64 --connect-back-address 127.0.0.1:9999` (there's way more switches, run with a -h to see these and override defaults)

This will then spit out the config key for the client. This hex string just includes the cert hash for TLS pinning, and the connect back address and port.

Then run up the client on the remote machine you want to pivot through, pick the right binary for the machine you're working with. In this case let's imagine it's netbsd on arm64:

`$ ./3gsocks_client_netbsd_arm64 505cb6d2460438313aa557f43ef0fefb5e414a5eeaabd6e340b5b6e4867d1cb53132372e302e302e313a39393939`

This will then connect back to your server and all going well anything you pipe down the local socks5 port will appear as if it's originating from the remote machine.

Any issues will end up pushed to stdout/stderr, so look at those if you're having issues.


#FAQ

Can I have multiple clients connect to the same server? Nooope

What about an android client? Try the linux binaries, and if that doesn't work then recompile using the android NDK.

It doesn't work!? I dunno duuude, fix it up and fire me a pull request.



## Acknowledgements

* The nspps golang rat authors who inspired me to replicate and improve their socks5 pivot.
* 3gstudent who's "homework" code I used as a base, as I have a sneaking (and totally unproven) suspicion was also used as a base by the nspps authors.

## License
BSD License, see LICENSE file