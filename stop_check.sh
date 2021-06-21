if [[ ! -e check_stream.pid ]]
then
	echo "not found"
else
	cat check_stream.pid | xargs kill
	rm -f check_stream.pid 
fi