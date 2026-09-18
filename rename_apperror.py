import os
import glob
import sys

def replace_in_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    new_content = content.replace('"github.com/alimtvnetwork/movie-cli-v8/apperror"', '"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"')
    
    # Also replace usage of apperror. as package qualifier
    new_content = new_content.replace('apperror.', 'appfault.')
    
    # Just in case it's used as a type `*apperror.AppError` -> `*appfault.AppError`
    # The previous replace handles it since it replaces `apperror.` with `appfault.`

    if new_content != content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(new_content)
        print(f"Updated {filepath}")
        return True
    return False

def main():
    updated_files = []
    # walk through all directories in D:\work\movie-cli-v8
    for root, dirs, files in os.walk(r'D:\work\movie-cli-v8'):
        if '.git' in root or '.ai-memory' in root or 'node_modules' in root:
            continue
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                if replace_in_file(filepath):
                    updated_files.append(filepath)
    
    with open('changed_files.txt', 'w') as f:
        for uf in updated_files:
            f.write(uf + '\n')

if __name__ == '__main__':
    main()
