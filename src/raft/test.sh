cat test_test.go| grep TestReElection2A | sed 's\(\ \g'|awk '/func/ {printf "%s ",$2;}' | xargs dstest -p 4 -o .run -v 1 -r  -s
