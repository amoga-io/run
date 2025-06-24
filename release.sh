#!/usr/bin/env bash
# release.sh - Enhanced helper script for creating releases with semantic versioning

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get current version
get_current_version() {
    git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
}

# Increment version
increment_version() {
    local version=$1
    local type=$2

    # Remove 'v' prefix if present
    version=${version#v}

    IFS='.' read -ra PARTS <<< "$version"
    major=${PARTS[0]}
    minor=${PARTS[1]}
    patch=${PARTS[2]}

    case $type in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch")
            patch=$((patch + 1))
            ;;
        *)
            print_error "Invalid version type. Use: major, minor, or patch"
            exit 1
            ;;
    esac

    echo "v$major.$minor.$patch"
}

# Show help
show_help() {
    echo -e "${BLUE}🚀 gocli Release Helper${NC}"
    echo ""
    echo "Usage: $0 <version-type> [message]"
    echo "   or: $0 <specific-version> [message]"
    echo ""
    echo -e "${YELLOW}Automatic versioning:${NC}"
    echo -e "  ${GREEN}patch${NC}  - Bug fixes ($(get_current_version) → $(increment_version "$(get_current_version)" "patch"))"
    echo -e "  ${GREEN}minor${NC}  - New features ($(get_current_version) → $(increment_version "$(get_current_version)" "minor"))"
    echo -e "  ${GREEN}major${NC}  - Breaking changes ($(get_current_version) → $(increment_version "$(get_current_version)" "major"))"
    echo ""
    echo -e "${YELLOW}Manual versioning:${NC}"
    echo -e "  ${GREEN}v1.2.3${NC} - Specific version"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 patch                           # Auto-increment patch"
    echo "  $0 minor \"Add new feature X\"       # Auto-increment minor with message"
    echo "  $0 major \"Breaking API changes\"    # Auto-increment major with message"
    echo "  $0 v2.1.0 \"Manual version\"         # Set specific version"
    echo ""
    echo -e "${BLUE}Current version: $(get_current_version)${NC}"
}

# Main script
main() {
    if [ $# -eq 0 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_help
        exit 0
    fi

    local version_input=$1
    local message=${2:-""}

    # Validate we're in a git repo
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi

    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        print_error "You have uncommitted changes. Please commit or stash them first."
        git status --short
        exit 1
    fi

    # Determine if input is version type or specific version
    local new_version
    if [[ "$version_input" =~ ^(patch|minor|major)$ ]]; then
        # Auto-increment version
        current_version=$(get_current_version)
        new_version=$(increment_version "$current_version" "$version_input")
        print_status "Auto-incrementing $version_input version"
    elif [[ "$version_input" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
        # Specific version provided
        new_version="$version_input"
        print_status "Using specific version"
    else
        print_error "Invalid input. Use 'patch', 'minor', 'major', or a specific version like 'v1.2.3'"
        echo ""
        show_help
        exit 1
    fi

    # Check if tag already exists
    if git tag -l | grep -q "^${new_version}$"; then
        print_error "Tag ${new_version} already exists"
        exit 1
    fi

    current_version=$(get_current_version)

    echo ""
    echo -e "${BLUE}📦 Release Information:${NC}"
    echo -e "Current version: ${YELLOW}$current_version${NC}"
    echo -e "New version: ${GREEN}$new_version${NC}"

    if [ -n "$message" ]; then
        echo -e "Message: ${YELLOW}$message${NC}"
    fi

    echo ""

    # Show recent commits since last tag
    echo -e "${BLUE}📝 Changes since $current_version:${NC}"
    if [ "$current_version" != "v0.0.0" ]; then
        git log --oneline --no-merges "$current_version"..HEAD | head -10 || true
    else
        git log --oneline --no-merges -10
    fi

    echo ""
    read -p "🚀 Proceed with release? (y/N): " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Release cancelled"
        exit 0
    fi

    # Create the release
    print_status "Creating release $new_version..."

    # Create tag with message
    if [ -n "$message" ]; then
        git tag -a "$new_version" -m "$message"
    else
        git tag -a "$new_version" -m "Release $new_version"
    fi

    print_success "✅ Created tag $new_version"

    # Build with new version
    print_status "🔨 Building with new version..."
    make build

    # Test the version
    print_status "🧪 Testing version command..."
    ./gocli version

    echo ""
    print_success "🎉 Release $new_version created successfully!"
    echo ""
    echo -e "${BLUE}📋 Next steps:${NC}"
    echo -e "  1. 🧪 Test thoroughly: ${YELLOW}./gocli --help${NC}"
    echo -e "  2. 📤 Push to remote: ${YELLOW}git push origin main --tags${NC}"
    echo -e "  3. 🌐 Create GitHub release (optional): ${YELLOW}https://github.com/amoga-io/run/releases/new${NC}"
    echo -e "  4. 📢 Users can update with: ${YELLOW}gocli update${NC}"
    echo ""
    echo -e "${GREEN}🔗 Direct install URL for users:${NC}"
    echo -e "${BLUE}bash <(curl -fsSL https://raw.githubusercontent.com/amoga-io/run/main/install.sh)${NC}"
}

main "$@"
