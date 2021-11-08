# Chainsaw

Circular buffer logging framework for Go. 

This logging library will keep the last N log messages around. This allows you to 
dump the logs at trace or debug level if you encounter an error, contextualizing
the error.


It also allows you to start a stream of log messages, making it suitable for 
applications which have sub-applications where you might wanna look at different 
streams of logs.


