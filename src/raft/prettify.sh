 for file in .run/1_.*.log ; do rm -rf file; done
 for file in .run/*.log ; do dslogs log -c 3 > $file.pretty; done
