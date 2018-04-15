#!/usr/bin/perl

# create a key with Go
$go = `go run example/create_key.go | gpg --no-default-keyring --list-packets`;

# create a key with GnuPG
`gpg --no-default-keyring --homedir /tmp/ --gen-key --batch <<EOF
%pubring /tmp/pubring.gpg
%secring /tmp/secring.gpg
%no-ask-passphrase
Key-Type: RSA
Key-Length: 2048
Key-Usage: sign
Subkey-Type: RSA
Subkey-Length: 2048
Subkey-Usage: encrypt
Expire-Date: 1y
Name-Real: JoeJoe
Name-Comment: test key
Name-Email: joe\@example.com
EOF`;
$gnupg = `cat /tmp/pubring.gpg /tmp/secring.gpg | gpg --no-default-keyring --list-packets`;

# To compare the output in $go and $gnupg, we need to:
# - parse the output and reorder entires
# - flatten the result
# - compare line-by-line

@go_list = parse_output($go);
@gnupg_list = parse_output($gnupg);

while (($#go_list > 0) && ($#gnupg_list > 0)) {
  $go_line = shift(@go_list);
  $gnupg_line = shift(@gnupg_list);

  if ($go_line ne $gnupg_line) {
    print "Go:    ", $go_line, "\n";
    print "GnuPG: ", $gnupg_line, "\n";
    die "found a difference!";
  }
}
# check if any rows left!
if ($#go_list > 0) {
  print join("\n", @go_list);
  die "Rows left in Go output!";
} elsif ($#gnupg_list > 0) {
  print join("\n", @gnupg_list);
  die "Rows left in GnuPG output!";
}

sub parse_output {
  my $output = shift(@_);
  my @r = ();
  while ($output =~ /(:[a-zA-Z ]+:)(.*?)(?=\n:|$)/sg) {
    push(@r, $1);
    @a = split(/\n/, $2);
    @b = ();
    foreach (@a) {
      $t = process_line($_);
      if ($t ne "") {
        push(@b, $t);
      }
    }
    @c = sort(@b);
    push(@r, @c);
  }
  return @r;
}

sub process_line {
  my $packet = shift(@_);
  $packet =~ s/^\s+//;

  # Handle data which changes at each run
  $packet =~ s/(keyid.*|key ID )[0-9A-F]{16}/$1/;
  $packet =~ s/created [0-9]+/created/;
  $packet =~ s/begin of digest [0-9a-f ]{5}/begin of digest/;
  $packet =~ s/checksum: [0-9a-f]{4}/checksum/;

  # Handle minor differences between GnuPG and Go
  $packet =~ s/len 3 \(pref-zip-algos: 2 3 1\)/len 2 (pref-zip-algos: 2 1)/; # Go doesn't support BZip2
  $packet =~ s/digest algo 2/digest algo 8/; # SHA1 vs SHA256

  # Extra stuff in Go
  $packet =~ s/hashed subpkt 25 len 1 \(primary user ID\)//;

  # Extra stuff in GnuPG
  $packet =~ s/hashed subpkt 23 len 1 \(key server preferences: 80\)//;
  $packet =~ s/hashed subpkt 30 len 1 \(features: 01\)//;

  # Ideally, we shouldn't need the following filters
  $packet =~ s/(data: )\[[0-9]{4} bits\]/$1/;
  $packet =~ s/(skey\[[0-9]\]: )\[[0-9]{4} bits\]/$1/;
  $packet =~ s/^subpkt/hashed subpkt/;
  return $packet;
}
