from hurry.filesize import size
from hurry.filesize import alternative

def convert_int_to_readable_file_size(file_size):
    return size(file_size, system=alternative).replace(' ', '')
