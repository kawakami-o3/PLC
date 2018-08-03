#!/usr/bin/env ruby

def build
	system "cd ../ && bash run.sh"
end

def run v

	if v.to_s[0] == '#' || v.to_s[0] == '('
		system "../scm2asm.exe '#{v}' > scheme_entry.s"
	else
		system "../scm2asm.exe #{v} > scheme_entry.s"
	end
	system "gcc -o a.out scheme_entry.s main.c"
	a = `./a.out`

	a.chomp!
	if a == v.to_s
		puts "OK: #{v}"
	else
		puts "NG: result #{a.chomp}, expected #{v}"
	end

	File.delete("a.out")
	File.delete("scheme_entry.s")
end

build

run 0
run 42
run 'a'
run 'z'
run 'M'
run '#t'
run '#f'
run '()'



