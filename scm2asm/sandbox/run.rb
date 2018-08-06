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
		puts "NG: result #{a.chomp}, expected #{result}"
		puts open("scheme_entry.s").read
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
run '(add1 3)', 4
run '(sub1 3)', 2
run '(integer->char 65)', 'A'
run '(char->integer A)', '65'
run '(zero? 1)', '#f'
run '(zero? 0)', '#t'
run '(null? ())', '#t'
run '(null? 1)', '#f'
run '(not #t)', '#f'
run '(not #f)', '#t'

run '(integer? 1)', '#t'
run '(integer? a)', '#f'
run '(integer? #t)', '#f'
run '(boolean? #t)', '#t'
run '(boolean? 8)', '#f'
run '(boolean? c)', '#f'



