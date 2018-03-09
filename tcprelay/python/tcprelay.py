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
            data = src.recv(16384)
            if not data:
                raise Exception("EOF")
            dst.send(data)
        except:
            break


def relay(local_conn, remoteaddr, remoteport):
    with closing(local_conn):
        try:
            remote_conn = socket.create_connection((remoteaddr, remoteport))
        except Exception as e:
            print("connect err: {} -> {}:{}".format(local_conn.getpeername(), remoteaddr, remoteport))
            return

        with closing(remote_conn):

            lp, rp = local_conn.getpeername(), remote_conn.getpeername()

            print("relay: {} <-> {}".format(lp, rp))

            done_queue = Queue()

            def f1(remote_conn, local_conn, done_queue, lp, rp):
                copy(remote_conn, local_conn)
                remote_conn.shutdown(socket.SHUT_WR)
                print("done: {} -> {}".format(lp, rp))
                done_queue.put(True)

            def f2(local_conn, remote_conn, done_queue, lp, rp):
                copy(local_conn, remote_conn)
                local_conn.shutdown(socket.SHUT_WR)
                print("done: {} <- {}".format(lp, rp))
                done_queue.put(True)

            t1 = threading.Thread(target=f1,
                                  args=(remote_conn, local_conn, done_queue, lp, rp))
            t2 = threading.Thread(target=f2,
                                  args=(local_conn, remote_conn, done_queue, lp, rp))

            t1.start()
            t2.start()

            for i in range(2):
                done_queue.get()


def main(arg):
    if len(arg) < 4:
        print("Usage:\n {} listenaddr listenport transimitaddr transimitport\n"
              "example:\n {} 0.0.0.0 10000 localhost 20000".format(arg[0], arg[0]))
        return

    listen_addr = arg[1]
    listen_port = int(arg[2])

    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_socket.bind((listen_addr, listen_port))
    server_socket.listen(5)

    while True:
        (client_socket, client_address) = server_socket.accept()

        t = threading.Thread(target=relay, args=(client_socket, arg[3], arg[4],))
        t.start()


if __name__ == '__main__':
    main(sys.argv)
