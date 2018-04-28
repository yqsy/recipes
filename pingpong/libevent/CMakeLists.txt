cmake_minimum_required(VERSION 3.9)
project(libevent)

set(CMAKE_CXX_STANDARD 11)

add_executable(client client.c)
target_link_libraries(client event_core)

add_executable(server server.c)
target_link_libraries(server event_core)
