# RFV video file format

File extention *.rvf

## Main structure

    <header>
    <metadata>
    <palette>
    <frames>

## Components

### header:

    str magic[3] = "RVF"
    u4 version    
    frame_size {
        u4 width
        u4 height
    }
    u4 frame_count
    f4 frame_time

This format version is "1", therefore first 4 bytes will be `(u4) 0x01465652`

### metadata:

    u4 metadata_size
    p1str title

`p1str` - pascal string with u1 size

### palette:

    u1 palette_size
    colors[palette_size] {
        u1 r,g,b
    }

### frames:

    <frame> frames[frame_count]

### frame:

    u4 frame_data_size  # incl. frame type & tail frame_data_size
    u1 frame_type
    <frame_data>
    u4 frame_data_size  # duplicate for backwards seeking

`frame_data_size` includes `frame_type` and tail `frame_data_size`

frame types:
- **RAW** - frame without compression
- **REPEAT** - frame repeating previous frame
- **SOLID** - frame filled with single color
- **ENCODED** - frame with comression
- **ENCODED_KEY** - frame with compression independent from previous frame
- **EOF** - endo of file marker

Frames RAW, SOLID and ENCODED_KEY are keyframes.

### frame_data

frame_type = RAW | ENCODED | ENCODED_KEY:

    u1 frame_data[frame_data_size - 1 - 4]

    if frame_type == RAW:
    frame_data_size = header.frame_size.width * header.frame_size.height + 1 + 4
    frame_data = u1 colors[header.frame_size.width * header.frame_size.height]
    

frame_type = SOLID:
    
    u1 r,g,b

    frame_data_size = 1 + 4 + 3
    
frame_type = REPEAT | EOF:
    
    empty

    frame_data_size = 1 + 4
    
    
### frame mapping:
    .. | prev_skip | next_skip | type | data | ...