# Run this script against the data in table A.1. on https://www.postgresql.org/docs/14/errcodes-appendix.html.
#
# Source data should be formatted like the following:
#
# Class 00 — Successful Completion
# 00000 	successful_completion
# Class 01 — Warning
# 01000 	warning
# 0100C 	dynamic_result_sets_returned
#
# for best results pass through gofmt
# ruby gen.rb < tablecontents.txt  | gofmt > errcode.go

code_name_overrides = {
  # Some error code names are repeated. In those cases add the error class as a suffix.
  "01004" => "StringDataRightTruncationWarning",
  "22001" => "StringDataRightTruncationDataException",
  "22004" => "NullValueNotAllowedDataException",
  "2F002" => "ModifyingSQLDataNotPermittedSQLRoutineException",
  "2F003" => "ProhibitedSQLStatementAttemptedSQLRoutineException",
  "2F004" => "ReadingSQLDataNotPermittedSQLRoutineException",
	"38002" => "ModifyingSQLDataNotPermittedExternalRoutineException",
	"38003" => "ProhibitedSQLStatementAttemptedExternalRoutineException",
  "38004" => "ReadingSQLDataNotPermittedExternalRoutineException",
  "39004" => "NullValueNotAllowedExternalRoutineInvocationException",

  # Go casing corrections
	"08001" => "SQLClientUnableToEstablishSQLConnection",
  "08004" => "SQLServerRejectedEstablishmentOfSQLConnection",
  "P0000" => "PLpgSQLError"
}

class_name_overrides = {
  # Go casing corrections
  "WITHCHECKOPTIONViolation" => "WithCheckOptionViolation"
}

cls_errs = Array.new
cls_assertions = Array.new
last_cls = ""
last_cls_full = ""

def build_assert_func(last_cls, last_cls_full, cls_errs)
  <<~GO
  // Is#{last_cls} asserts the error code class is #{last_cls_full}
  func Is#{last_cls} (code string) bool {
      switch code{
          case #{cls_errs.join(", ")}:
              return true
      }
      return false
  }
  GO
end

puts <<~STR
// Package pgerrcode contains constants for PostgreSQL error codes.
package pgerrcode

// Source: https://www.postgresql.org/docs/14/errcodes-appendix.html
// See gen.rb for script that can convert the error code table to Go code.

const (
STR

ARGF.each do |line|
  case line
  when /^Class/
    if cls_errs.length > 0 && last_cls != ""
      assert_func = build_assert_func(class_name_overrides.fetch(last_cls) { last_cls }, last_cls_full, cls_errs)
      cls_assertions.push(assert_func)
    end
    last_cls = line.split("—")[1]
    .gsub(" ", "")
    .gsub("/", "")
    .gsub("\n", "")
    .sub(/\(\w*\)/, "")
    last_cls_full = line.gsub("\n", "")
    cls_errs.clear
    puts
    puts "// #{line}"
  when /^(\w{5})\s+(\w+)/
    code = $1
    name = code_name_overrides.fetch(code) do
      $2.split("_").map(&:capitalize).join
        .gsub("Sql", "SQL")
        .gsub("Xml", "XML")
        .gsub("Fdw", "FDW")
        .gsub("Srf", "SRF")
        .gsub("Io", "IO")
        .gsub("Json", "JSON")
    end
    cls_errs.push(name)
    puts %Q[#{name} = "#{code}"]
  else
    puts line
  end
end
puts ")"

if cls_errs.length > 0
  assert_func = build_assert_func(class_name_overrides.fetch(last_cls) { last_cls }, last_cls_full, cls_errs)
  cls_assertions.push(assert_func)
end

cls_assertions.each do |cls_assertion|
  puts cls_assertion
end
