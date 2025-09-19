#!/bin/bash

# Script to update all provider files with standardized SAI functions
# This replaces old function calls with the new standardized format

echo "Updating provider files with standardized SAI functions..."

# Function to update a provider file
update_provider() {
    local file="$1"
    echo "Updating $file..."
    
    # Backup original file
    cp "$file" "$file.bak"
    
    # Replace old function calls with standardized ones
    sed -i.tmp \
        -e 's/{{sai_packages(\([^}]*\))}}/{{sai_package('\''*'\'', '\''name'\'')}}/g' \
        -e 's/{{sai_package(\([^}]*\))}}/{{sai_package(0, '\''name'\'')}}/g' \
        -e 's/{{sai_service(\([^}]*\))}}/{{sai_service(0, '\''service_name'\'')}}/g' \
        -e 's/{{sai_port()}}/{{sai_port(0, '\''port'\'')}}/g' \
        -e 's/{{sai_file()}}/{{sai_file(0, '\''path'\'')}}/g' \
        -e 's/{{sai_command(\([^}]*\))}}/{{sai_command(0, '\''path'\'')}}/g' \
        "$file"
    
    # Remove temporary file
    rm -f "$file.tmp"
}

# Update all provider files
for provider_file in providers/*.yaml; do
    if [ -f "$provider_file" ]; then
        update_provider "$provider_file"
    fi
done

# Update specialized providers
for provider_file in providers/specialized/*.yaml; do
    if [ -f "$provider_file" ]; then
        update_provider "$provider_file"
    fi
done

echo "Provider update complete!"
echo "Backup files created with .bak extension"