#!/bin/bash
# Script to create a demo Terraform project in /demo/ with examples from /examples/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$REPO_ROOT/.env"
DEMO_DIR="$REPO_ROOT/demo"
TEMPLATE_DIR="$SCRIPT_DIR/template"

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: .env file not found at $ENV_FILE" >&2
    echo "Please create it from .env-example" >&2
    exit 1
fi

# Remove existing demo directory if it exists
if [ -d "$DEMO_DIR" ]; then
    echo "Removing existing demo directory..."
    rm -rf "$DEMO_DIR"
fi

# Create demo directory
mkdir -p "$DEMO_DIR"

echo "Creating demo project in $DEMO_DIR..."

# Generate provider.tf directly into the demo directory using values from .env
echo "Generating provider.tf..."
OUTPUT_FILE="$DEMO_DIR/provider.tf" "$SCRIPT_DIR/generate-provider-tf.sh"

# Create versions.tf for local development (without version constraint)
echo "Creating versions.tf..."
cat > "$DEMO_DIR/versions.tf" <<'EOF'
terraform {
  required_providers {
    privx = {
      source = "registry.terraform.io/hashicorp/privx"
    }
  }
}
EOF

# Copy selected examples - good candidates for a demo
echo "Copying example files..."

# Define which examples are used (for documentation)
EXAMPLES_USED=(
  "examples/resources/privx_access_group/resource.tf -> resource-access-group.tf"
  "examples/resources/privx_whitelist/resource.tf -> resource-whitelist.tf"
  "examples/resources/privx_host/resource.tf -> resource-host.tf"
  "examples/resources/privx_role/resource.tf -> resource-role.tf"
)

# Resources (will create actual resources) - all have no dependencies
cp "$REPO_ROOT/examples/resources/privx_host/resource.tf" "$DEMO_DIR/resource-host.tf"
cp "$REPO_ROOT/examples/resources/privx_role/resource.tf" "$DEMO_DIR/resource-role.tf"
cp "$REPO_ROOT/examples/resources/privx_access_group/resource.tf" "$DEMO_DIR/resource-access-group.tf"
cp "$REPO_ROOT/examples/resources/privx_whitelist/resource.tf" "$DEMO_DIR/resource-whitelist.tf"

# Copy outputs.tf from template
echo "Creating outputs.tf..."
OUTPUTS_TEMPLATE="$TEMPLATE_DIR/outputs.tf.tf-demo"
if [ ! -f "$OUTPUTS_TEMPLATE" ]; then
    echo "Error: outputs.tf template not found at $OUTPUTS_TEMPLATE" >&2
    exit 1
fi
cp "$OUTPUTS_TEMPLATE" "$DEMO_DIR/outputs.tf"

# Copy .gitignore from template
echo "Creating .gitignore..."
GITIGNORE_TEMPLATE="$TEMPLATE_DIR/.gitignore.tf-demo"
if [ ! -f "$GITIGNORE_TEMPLATE" ]; then
    echo "Error: .gitignore template not found at $GITIGNORE_TEMPLATE" >&2
    exit 1
fi
cp "$GITIGNORE_TEMPLATE" "$DEMO_DIR/.gitignore"

# Copy README from template
echo "Creating README.md..."
README_TEMPLATE="$TEMPLATE_DIR/README.md.tf-demo"
if [ ! -f "$README_TEMPLATE" ]; then
    echo "Error: README template not found at $README_TEMPLATE" >&2
    exit 1
fi
cp "$README_TEMPLATE" "$DEMO_DIR/README.md"

echo "Created demo project at: $DEMO_DIR"
echo ""
echo "Files created:"
echo "  - versions.tf (provider requirements for local development)"
echo "  - provider.tf (generated from .env values)"
echo "  - resource-access-group.tf (from examples/resources/privx_access_group/resource.tf)"
echo "  - resource-whitelist.tf (from examples/resources/privx_whitelist/resource.tf)"
echo "  - resource-host.tf (from examples/resources/privx_host/resource.tf)"
echo "  - resource-role.tf (from examples/resources/privx_role/resource.tf)"
echo "  - outputs.tf (references the above resources)"
echo ""
echo "Source examples used:"
for example in "${EXAMPLES_USED[@]}"; do
  echo "  - $example"
done
echo ""
echo "Next steps:"
echo "1. cd $DEMO_DIR"
echo "2. (Optional) Update resource names in the .tf files to customize them"
echo "3. terraform plan (skip 'terraform init' - not needed with dev_overrides)"
echo "4. terraform apply"
echo ""
echo "Note: All resources have no dependencies and can be created independently!"
echo ""
echo "To modify the demo, edit the source files in examples/ and run 'task demo' again."

