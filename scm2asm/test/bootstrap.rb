#!/usr/bin/env ruby

require "open3"

def build
	system "cd ../ && goimports -w -l . && go build -o scm2asm.exe"
end

def run v, result = nil
	result ||= v
	result = result.to_s

	if v.to_s[0] == '#' || v.to_s[0] == '('
		system "../scm2asm.exe -- '#{v}' > scheme_entry.s"
	else
		system "../scm2asm.exe -- #{v} > scheme_entry.s"
	end
	system "gcc -o a.out scheme_entry.s main.c"
	#a = `./a.out`
	o, e, s = Open3.capture3("./a.out")

	o.chomp!
	if o == result
		puts "OK: #{v}"
	else
		printf("NG: #{v}\n")
		printf("Result: #{o.chomp} (%#b), expected #{result}\n", o.chomp.to_i)
		puts open("scheme_entry.s").read
		p [o,e,s]
		exit
	end

	#File.delete("a.out")
	#File.delete("scheme_entry.s")
end

unless $done
	build
	$done
end



def run_all name, tests
	puts "Test: #{name}"
	tests.each do |k,v|
		run k,v
	end
end
