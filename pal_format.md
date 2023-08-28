# Palette file format

File extention *.pal

Filesize determines palette size.

All colors are sorted by luma from darkest to lightest.

## Format

    file:
        <color> colors[file_size / 3]

    color:
        u1 r,g,b