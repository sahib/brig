#!/usr/bin/perl

# create a keyring with Go
$go = `go run example/create_key.go`;

# check that we can use it
`echo "hello world" | gpg --trust-model always --no-default-keyring --homedir /tmp/ --keyring ./pub.gpg --secret-keyring ./priv.gpg -a -e -r joe > cipher.txt 2>/dev/null`;

$output = `cat cipher.txt | gpg --no-tty --no-default-keyring --homedir /tmp/ --keyring ./pub.gpg --secret-keyring ./priv.gpg -d -u joe 2>/dev/null`;

if ($output != "hello world") {
  die("expecting 'hello world', got: " . $output)
}
print($output);
