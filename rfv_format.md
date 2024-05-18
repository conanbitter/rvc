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
	u1 flags

This format version is "2", therefore first 4 bytes will be `(u4) 0x02465652`

Flags may be:
|Flag|Value|
|---|---|
|COMPRESSION_NONE|0b00000000|
|COMPRESSION_FULL|0b00000001|

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

### frame (for uncompressed files):

    u1 frame_data[width * height]

### frame (for compressed files):

    u4 frame_data_size  # incl. flags & tail frame_data_size
    u1 flags
    u1 frame_data[frame_data_size - 1 - 4]
    u4 frame_data_size  # duplicate for backwards seeking

`frame_data_size` includes `flags` and tail `frame_data_size`

Flags may be:
|Flag|Value|Description
|---|---|---|
|IS_REGULAR|0b00000000|This is a regular frame|
|IS_KEYFRAME|0b00000001|This is a keyframe (independent from a previous frame)
|IS_FIRST|0b00000010|This is the first frame in file
|IS_LAST|0b00000100|This is the last frame in file

### frame mapping:
    .. | prev_skip | next_skip | flags | data | ...