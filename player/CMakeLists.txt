cmake_minimum_required(VERSION 3.16)

set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} ${CMAKE_CURRENT_SOURCE_DIR}/cmake)

project(player VERSION 0.1 LANGUAGES C)

set(EXECUTABLE_OUTPUT_PATH ${CMAKE_CURRENT_SOURCE_DIR}/bin)

add_executable(player
    src/rvf_decode.h
    src/rvf_decode.c
    src/decoding.h
    src/decoding.c
    src/main.c
)

find_package(SDL2 REQUIRED)

if(NOT CMAKE_BUILD_TYPE STREQUAL "Debug")
    set_property(TARGET player PROPERTY MSVC_RUNTIME_LIBRARY "MultiThreaded")
    target_link_options(player PRIVATE -SUBSYSTEM:WINDOWS)
    target_compile_definitions(player PRIVATE "NOCONSOLE=1")
endif()

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