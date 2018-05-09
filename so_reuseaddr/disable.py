import socket
import sys
import threading
import os

# 做成discard吧
def serve(client_socket):
    while(True):
        data = os.read(client_socket.fileno(), 16384)
        if not data:
            break
    
    print("socket close")

def main(argv):
    if len(argv) < 3:
        print("Usage:\n %s listenIp listenPort\n" % (argv[0]))
        return
    
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.bind((argv[1], int(argv[2])))
    server_socket.listen(5)
    
    while(True):
        (clinet_socket, client_address) = server_socket.accept()
        print("{} accepted".format(client_address))

        tr = threading.Thread(target=serve,args=(clinet_socket,))
        tr.start()
        # detach tr

if __name__ == '__main__':
    main(sys.argv)
