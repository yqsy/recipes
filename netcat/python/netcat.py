#!/usr/bin/python3

import socket
import sys
import threading
from contextlib import contextmanager
from queue import Queue

import os


@contextmanager
def closing(conn):
    try:
        yield conn
    finally:
        conn.close()


def copy(dst, src):
    while True:
        try:
            data = os.read(src, 16384)
            if not data:
                raise Exception("EOF")
            os.write(dst, data)
        except:
            break


def relay(conn):
    with closing(conn):
        done_queue = Queue()

        active = 1
        passive = 2

        # [stdin] -> remote
        def f1(conn, done_queue):
            copy(conn.fileno(), sys.stdin.fileno())
            done_queue.put(active)

        # stdout <- [remote]
        def f2(conn, done_queue):
            copy(sys.stdout.fileno(), conn.fileno())
            done_queue.put(passive)

        t1 = threading.Thread(target=f1, args=(conn, done_queue,))
        t2 = threading.Thread(target=f2, args=(conn, done_queue,))

        t1.start()
        t2.start()

        first = done_queue.get()

        if first == active:
            conn.shutdown(socket.SHUT_WR)
            done_queue.get()
        else:
            # how to stop read from stdin?
            # done_queue.get()
            # conn.shutdown(socket.SHUT_WR)
            os._exit(0)


def main(argv):
    if len(argv) < 3:
        print("Usage:\n  %s -l port\n  %s host port\n" % (argv[0], argv[0]))
        return

    port = int(argv[2])
    if argv[1] == "-l":
        # server
        server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        server_socket.bind(('', port))
        server_socket.listen(5)
        (client_socket, client_address) = server_socket.accept()
        server_socket.close()
        relay(client_socket)
    else:
        # client
        sock = socket.create_connection((argv[1], port))
        relay(sock)


if __name__ == '__main__':
    main(sys.argv)
