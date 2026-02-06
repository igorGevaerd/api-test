#!/bin/bash

# Deploy script for Terraform ECS Fargate Infrastructure
# Usage: ./deploy.sh [dev|prod] [plan|apply|destroy]

set -e

ENVIRONMENT=${1:-dev}
ACTION=${2:-plan}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Validate arguments
if [[ ! "$ENVIRONMENT" =~ ^(dev|prod)$ ]]; then
    echo -e "${RED}❌ Invalid environment. Use 'dev' or 'prod'${NC}"
    exit 1
fi

if [[ ! "$ACTION" =~ ^(plan|apply|destroy|output|state)$ ]]; then
    echo -e "${RED}❌ Invalid action. Use 'plan', 'apply', 'destroy', 'output', or 'state'${NC}"
    exit 1
fi

# Set working directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
WORK_DIR="$SCRIPT_DIR/environments/$ENVIRONMENT"

echo -e "${GREEN}🚀 Terraform Deployment Script${NC}"
echo -e "${YELLOW}Environment: $ENVIRONMENT${NC}"
echo -e "${YELLOW}Action: $ACTION${NC}"
echo ""

# Function to initialize Terraform
init_terraform() {
    echo -e "${YELLOW}📦 Initializing Terraform...${NC}"
    cd "$WORK_DIR"
    terraform init
    cd - > /dev/null
    echo -e "${GREEN}✅ Terraform initialized${NC}"
}

# Function to validate configuration
validate_terraform() {
    echo -e "${YELLOW}🔍 Validating Terraform configuration...${NC}"
    cd "$WORK_DIR"
    terraform validate
    cd - > /dev/null
    echo -e "${GREEN}✅ Configuration is valid${NC}"
}

# Function to format code
format_terraform() {
    echo -e "${YELLOW}🎨 Formatting Terraform code...${NC}"
    cd "$SCRIPT_DIR"
    terraform fmt -recursive
    cd - > /dev/null
    echo -e "${GREEN}✅ Code formatted${NC}"
}

# Function to plan deployment
plan_terraform() {
    echo -e "${YELLOW}📋 Planning Terraform deployment...${NC}"
    cd "$WORK_DIR"
    terraform plan -out=tfplan
    cd - > /dev/null
    echo -e "${GREEN}✅ Plan complete. Review above and run with 'apply' action${NC}"
}

# Function to apply deployment
apply_terraform() {
    echo -e "${YELLOW}⚙️  Applying Terraform configuration...${NC}"
    cd "$WORK_DIR"
    
    if [ ! -f tfplan ]; then
        echo -e "${YELLOW}No tfplan found. Creating new plan...${NC}"
        terraform plan -out=tfplan
    fi
    
    terraform apply tfplan
    rm -f tfplan
    cd - > /dev/null
    echo -e "${GREEN}✅ Deployment complete${NC}"
}

# Function to destroy infrastructure
destroy_terraform() {
    echo -e "${RED}⚠️  WARNING: This will destroy all infrastructure in $ENVIRONMENT${NC}"
    read -p "Type 'destroy' to confirm: " confirmation
    
    if [ "$confirmation" != "destroy" ]; then
        echo "Cancelled"
        exit 0
    fi
    
    echo -e "${YELLOW}🔥 Destroying Terraform infrastructure...${NC}"
    cd "$WORK_DIR"
    terraform destroy
    cd - > /dev/null
    echo -e "${GREEN}✅ Infrastructure destroyed${NC}"
}

# Function to show outputs
show_outputs() {
    echo -e "${YELLOW}📊 Terraform Outputs:${NC}"
    cd "$WORK_DIR"
    terraform output
    cd - > /dev/null
}

# Function to show state
show_state() {
    echo -e "${YELLOW}📊 Terraform State:${NC}"
    cd "$WORK_DIR"
    terraform state list
    cd - > /dev/null
}

# Main execution
main() {
    init_terraform
    validate_terraform
    format_terraform
    
    case "$ACTION" in
        plan)
            plan_terraform
            ;;
        apply)
            plan_terraform
            apply_terraform
            ;;
        destroy)
            destroy_terraform
            ;;
        output)
            show_outputs
            ;;
        state)
            show_state
            ;;
    esac
}

main
