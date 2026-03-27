import os

# Cartelle da ignorare assolutamente
IGNORE_DIRS = {'node_modules', '.git', 'vendor', 'dist', 'build', '.next'}
# Estensioni utili
ALLOWED_EXT = {'.go', '.js', '.jsx', '.ts', '.tsx', '.json', '.html', '.css'}

def bundle_project(root_dir, output_file):
    with open(output_file, 'w', encoding='utf-8') as f_out:
        for root, dirs, files in os.walk(root_dir):
            # Filtra le cartelle da ignorare
            dirs[:] = [d for d in dirs if d not in IGNORE_DIRS]
            
            for file in files:
                if any(file.endswith(ext) for ext in ALLOWED_EXT):
                    if file in ['package-lock.json', 'go.sum']: continue # Troppo lunghi e inutili
                    
                    file_path = os.path.join(root, file)
                    f_out.write(f"\n\n{'='*20}\nPATH: {file_path}\n{'='*20}\n")
                    try:
                        with open(file_path, 'r', encoding='utf-8') as f_in:
                            f_out.write(f_in.read())
                    except Exception as e:
                        f_out.write(f"[Errore lettura file: {e}]")

bundle_project('.', 'full_context_project.txt')
print("File generato: full_context_project.txt")
