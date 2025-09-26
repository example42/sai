#!/bin/bash

# SAI Data Validation Script
# Validates all saidata files against the schema

set -e

SCHEMA_FILE="schemas/saidata-0.2-schema.json"
SAMPLES_DIR="docs/saidata_samples"
FAILED_FILES=()
TOTAL_FILES=0
VALID_FILES=0

echo "🔍 SAI Data Schema Validation"
echo "=============================="
echo "Schema: $SCHEMA_FILE"
echo "Samples: $SAMPLES_DIR"
echo ""

# Check if ajv-cli is available
if ! command -v ajv &> /dev/null; then
    echo "❌ ajv-cli not found. Installing..."
    npm install -g ajv-cli
fi

# Check if schema file exists
if [[ ! -f "$SCHEMA_FILE" ]]; then
    echo "❌ Schema file not found: $SCHEMA_FILE"
    exit 1
fi

# Check if samples directory exists
if [[ ! -d "$SAMPLES_DIR" ]]; then
    echo "❌ Samples directory not found: $SAMPLES_DIR"
    exit 1
fi

echo "📋 Validating saidata files..."
echo ""

# Find and validate all YAML files
while IFS= read -r -d '' file; do
    TOTAL_FILES=$((TOTAL_FILES + 1))
    relative_path="${file#./}"
    
    echo -n "  $(basename "$file") ($(dirname "$relative_path"))... "
    
    if ajv validate -s "$SCHEMA_FILE" -d "$file" >/dev/null 2>&1; then
        echo "✅"
        VALID_FILES=$((VALID_FILES + 1))
    else
        echo "❌"
        FAILED_FILES+=("$file")
    fi
done < <(find "$SAMPLES_DIR" -name "*.yaml" -print0)

echo ""
echo "📊 Validation Summary"
echo "===================="
echo "Total files: $TOTAL_FILES"
echo "Valid files: $VALID_FILES"
echo "Failed files: $((TOTAL_FILES - VALID_FILES))"

if [[ ${#FAILED_FILES[@]} -gt 0 ]]; then
    echo ""
    echo "❌ Failed Files:"
    for file in "${FAILED_FILES[@]}"; do
        echo "  - $file"
        echo "    Error details:"
        ajv validate -s "$SCHEMA_FILE" -d "$file" 2>&1 | sed 's/^/      /'
        echo ""
    done
    exit 1
else
    echo ""
    echo "🎉 All saidata files are valid!"
    exit 0
fi