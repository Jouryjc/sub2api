#!/usr/bin/env python3
import socket
import sys
import threading


LISTEN_HOST = "172.19.0.1"
LISTEN_PORT = 7892
TARGET_HOST = "127.0.0.1"
TARGET_PORT = 7890


def close_socket(sock):
    try:
        sock.shutdown(socket.SHUT_RDWR)
    except Exception:
        pass
    try:
        sock.close()
    except Exception:
        pass


def relay(src, dst):
    try:
        while True:
            data = src.recv(65536)
            if not data:
                break
            dst.sendall(data)
    except Exception:
        pass
    finally:
        close_socket(src)
        close_socket(dst)


def handle(client):
    try:
        upstream = socket.create_connection((TARGET_HOST, TARGET_PORT), timeout=10)
    except Exception:
        close_socket(client)
        return
    threading.Thread(target=relay, args=(client, upstream), daemon=True).start()
    threading.Thread(target=relay, args=(upstream, client), daemon=True).start()


def main():
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind((LISTEN_HOST, LISTEN_PORT))
    server.listen(256)
    print(f"forwarding {LISTEN_HOST}:{LISTEN_PORT} -> {TARGET_HOST}:{TARGET_PORT}", flush=True)
    while True:
        client, _ = server.accept()
        threading.Thread(target=handle, args=(client,), daemon=True).start()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit(0)
