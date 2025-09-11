#!/bin/bash

echo "🔍 Checking for business logic leaks in DevTools packages..."
echo ""

# Check for Skyscape-specific references
echo "1. Checking for Skyscape references..."
SKYSCAPE_REFS=$(grep -r "Skyscape\|skyscape\|theskyscape\|skysca\.pe" pkg/emailing pkg/payments --include="*.go" 2>/dev/null | grep -v "github.com/The-Skyscape" | wc -l)
if [ $SKYSCAPE_REFS -eq 0 ]; then
    echo "   ✅ No Skyscape-specific references found"
else
    echo "   ❌ Found $SKYSCAPE_REFS Skyscape references"
    grep -r "Skyscape\|skyscape\|theskyscape\|skysca\.pe" pkg/emailing pkg/payments --include="*.go" | grep -v "github.com/The-Skyscape"
fi

# Check for business-specific terms
echo ""
echo "2. Checking for business-specific terms..."
BUSINESS_TERMS=$(grep -r "workspace\|Workspace\|workbench\|Workbench\|commander\|Commander" pkg/emailing pkg/payments --include="*.go" -i 2>/dev/null | wc -l)
if [ $BUSINESS_TERMS -eq 0 ]; then
    echo "   ✅ No business-specific terms found"
else
    echo "   ❌ Found $BUSINESS_TERMS business-specific terms"
    grep -r "workspace\|Workspace\|workbench\|Workbench\|commander\|Commander" pkg/emailing pkg/payments --include="*.go" -i
fi

# Check for hardcoded business URLs
echo ""
echo "3. Checking for hardcoded business URLs..."
BUSINESS_URLS=$(grep -r "theskyscape\.com\|skyscape\.dev\|skysca\.pe" pkg/emailing pkg/payments --include="*.go" 2>/dev/null | wc -l)
if [ $BUSINESS_URLS -eq 0 ]; then
    echo "   ✅ No hardcoded business URLs found"
else
    echo "   ❌ Found $BUSINESS_URLS business URLs"
    grep -r "theskyscape\.com\|skyscape\.dev\|skysca\.pe" pkg/emailing pkg/payments --include="*.go"
fi

# Check for specific product names
echo ""
echo "4. Checking for specific product names..."
PRODUCT_NAMES=$(grep -r "Student Workbench\|Workspace Pro" pkg/emailing pkg/payments --include="*.go" 2>/dev/null | wc -l)
if [ $PRODUCT_NAMES -eq 0 ]; then
    echo "   ✅ No specific product names found"
else
    echo "   ❌ Found $PRODUCT_NAMES specific product names"
    grep -r "Student Workbench\|Workspace Pro" pkg/emailing pkg/payments --include="*.go"
fi

# Check email templates for business content
echo ""
echo "5. Checking email templates for business content..."
if [ -d "pkg/emailing/views" ]; then
    TEMPLATE_REFS=$(grep -r "Skyscape\|workspace\|workbench" pkg/emailing/views --include="*.html" 2>/dev/null | wc -l)
    if [ $TEMPLATE_REFS -eq 0 ]; then
        echo "   ✅ No business content in email templates"
    else
        echo "   ❌ Found $TEMPLATE_REFS business references in templates"
        grep -r "Skyscape\|workspace\|workbench" pkg/emailing/views --include="*.html"
    fi
else
    echo "   ✅ No email template directory found (good - templates should be in apps)"
fi

# Summary
echo ""
echo "📊 Summary:"
if [ $SKYSCAPE_REFS -eq 0 ] && [ $BUSINESS_TERMS -eq 0 ] && [ $BUSINESS_URLS -eq 0 ] && [ $PRODUCT_NAMES -eq 0 ] && [ $TEMPLATE_REFS -eq 0 ]; then
    echo "   ✅ DevTools packages are clean - no business logic leaks detected!"
else
    echo "   ⚠️  Found potential business logic leaks - please review above"
fi