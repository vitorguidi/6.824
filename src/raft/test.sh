cat test_test.go| grep 2A | sed 's\(\ \g'|awk '/func/ {printf "%s ",$2;}' | xargs dstest -p 8 -o .run -v 1 -s -n 100
