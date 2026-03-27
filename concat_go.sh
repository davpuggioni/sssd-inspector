#!/bin/bash

# Nome del file di output
OUTPUT="project_context.txt"

# Svuota il file se esiste già
> $OUTPUT

echo "--- INIZIO STRUTTURA PROGETTO ---" >> $OUTPUT
tree -I 'vendor|node_modules|.git' >> $OUTPUT
echo -e "--- FINE STRUTTURA PROGETTO ---\n" >> $OUTPUT

# Trova tutti i file .go, escludendo la cartella vendor
find . -name "*.go" -not -path "./vendor/*" | while read -r file; do
    echo "========================================" >> $OUTPUT
    echo "FILE: $file" >> $OUTPUT
    echo "========================================" >> $OUTPUT
    cat "$file" >> $OUTPUT
    echo -e "\n\n" >> $OUTPUT
done

echo "Fatto! Carica '$OUTPUT' su Gemini."
