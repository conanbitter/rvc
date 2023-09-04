cmake_minimum_required(VERSION 3.16)

set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} ${CMAKE_CURRENT_SOURCE_DIR}/cmake)

project(player VERSION 0.1 LANGUAGES C)

set(EXECUTABLE_OUTPUT_PATH ${CMAKE_CURRENT_SOURCE_DIR}/bin)

add_executable(player
    src/rvf_decode.h
    src/rvf_decode.c
    src/main.c
)

find_package(SDL2 REQUIRED)

target_include_directories(player
    PUBLIC
        ${SDL2_INCLUDE_DIR}    
)

target_link_directories(player
    PUBLIC
        ${SDL2_LIBRARY}
)

target_link_libraries(player
    SDL2
)