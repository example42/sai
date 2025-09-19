#!/bin/bash

# Script to update all provider files with the provider parameter in SAI functions
# This adds the third parameter to all SAI function calls

echo "Adding provider parameter to all SAI functions in provider files..."

# Function to update a provider file with its provider name
update_provider_with_param() {
    local file="$1"
    local provider_name="$2"
    
    echo "Updating $file with provider parameter '$provider_name'..."
    
    # Backup original file
    cp "$file" "$file.bak"
    
    # Replace SAI function calls to include provider parameter
    sed -i.tmp \
        -e "s/{{sai_package(\([^,}]*\), *'\([^']*\)')}}/{{sai_package(\1, '\2', '$provider_name')}}/g" \
        -e "s/{{sai_service(\([^,}]*\), *'\([^']*\)')}}/{{sai_service(\1, '\2', '$provider_name')}}/g" \
        -e "s/{{sai_file(\([^,}]*\), *'\([^']*\)')}}/{{sai_file(\1, '\2', '$provider_name')}}/g" \
        -e "s/{{sai_directory(\([^,}]*\), *'\([^']*\)')}}/{{sai_directory(\1, '\2', '$provider_name')}}/g" \
        -e "s/{{sai_command(\([^,}]*\), *'\([^']*\)')}}/{{sai_command(\1, '\2', '$provider_name')}}/g" \
        -e "s/{{sai_port(\([^,}]*\), *'\([^']*\)')}}/{{sai_port(\1, '\2', '$provider_name')}}/g" \
        -e "s/{{sai_container(\([^,}]*\), *'\([^']*\)')}}/{{sai_container(\1, '\2', '$provider_name')}}/g" \
        "$file"
    
    # Remove temporary file
    rm -f "$file.tmp"
}

# Extract provider name from YAML file
get_provider_name() {
    local file="$1"
    grep -E "^\s*name:\s*\".*\"" "$file" | sed 's/.*name: *"\([^"]*\)".*/\1/' | head -1
}

# Update all provider files
for provider_file in providers/*.yaml; do
    if [ -f "$provider_file" ]; then
        provider_name=$(get_provider_name "$provider_file")
        if [ -n "$provider_name" ]; then
            update_provider_with_param "$provider_file" "$provider_name"
        else
            echo "Warning: Could not extract provider name from $provider_file"
        fi
    fi
done

# Update specialized providers
for provider_file in providers/specialized/*.yaml; do
    if [ -f "$provider_file" ]; then
        provider_name=$(get_provider_name "$provider_file")
        if [ -n "$provider_name" ]; then
            update_provider_with_param "$provider_file" "$provider_name"
        else
            echo "Warning: Could not extract provider name from $provider_file"
        fi
    fi
done

echo "Provider parameter update complete!"
echo "Backup files created with .bak extension"