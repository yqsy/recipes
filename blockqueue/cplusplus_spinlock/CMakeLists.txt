cmake_minimum_required(VERSION 3.9)
project(cplusplus_spinlock)

set(CMAKE_CXX_STANDARD 11)

add_executable(cplusplus_spinlock main.cpp)
target_link_libraries(cplusplus_spinlock pthread)
