#!/bin/bash
exec egrep -h '^func \(.+?\) [A-Z][_a-zA-Z0-9]*\(.+?\).+? {$' "$@" \
     | perl -e '
%data=();
while(<>){
  chomp($_);
  if ($_ =~ m!^\s*func \(\w+ (.+?)\) ([_a-zA-Z0-9]+.+?)\s*\{\s*$!) {
    ($s, $m) = ($1, $2);
    $data{$s} ||= [];
    push(@{$data{$s}},$m);
  } else {
    print STDERR "error parsing func string: ".$_."\n";
  }
};
$first=1;
foreach $k (sort keys %data) {
  print "\n" unless $first;
  $first=0 if $first;
  print "// type: ".$k."\n";
  @unique = do { my %seen; grep { !$seen{$_}++ } @{$data{$k}} };
  #@sorted = sort {
  #  if (length($a) == length($b)) {
  #    return $a <=> $b;
  #  }
  #  return length($a) <=> length($b);
  #} @unique;
  @sorted = @unique;
  foreach $m (@sorted) {
    print $m."\n";
  }
};
'
