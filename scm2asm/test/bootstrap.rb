#!/usr/bin/env ruby

def build
	system "cd ../ && bash run.sh"
end

def run v, result = nil
	result ||= v
	result = result.to_s

	if v.to_s[0] == '#' || v.to_s[0] == '('
		system "../scm2asm.exe '#{v}' > scheme_entry.s"
	else
		system "../scm2asm.exe #{v} > scheme_entry.s"
	end
	system "gcc -o a.out scheme_entry.s main.c"
	a = `./a.out`

	a.chomp!
	if a == result
		puts "OK: #{v}"
	else
		printf("NG: result #{a.chomp} (%#b), expected #{result}\n", a.chomp.to_i)
		puts open("scheme_entry.s").read
	end

	#File.delete("a.out")
	#File.delete("scheme_entry.s")
end

build



