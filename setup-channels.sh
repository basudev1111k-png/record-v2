#!/bin/bash
# Simple setup script for GoondVR channels configuration

echo "🎬 GoondVR Channel Setup"
echo "========================"
echo ""

# Create conf directory if it doesn't exist
if [ ! -d "conf" ]; then
    echo "📁 Creating conf directory..."
    mkdir -p conf
fi

# Check if channels.json already exists
if [ -f "conf/channels.json" ]; then
    echo "⚠️  conf/channels.json already exists!"
    echo ""
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Setup cancelled. Your existing channels.json was not modified."
        exit 0
    fi
fi

# Copy example file
if [ -f "channels.json.example" ]; then
    echo "📋 Copying channels.json.example to conf/channels.json..."
    cp channels.json.example conf/channels.json
    echo "✅ Done!"
    echo ""
    echo "📝 Next steps:"
    echo "   1. Edit conf/channels.json and add your channels"
    echo "   2. Run ./goondvr to start recording"
    echo ""
    echo "📖 For detailed documentation, see CHANNELS.md"
else
    echo "❌ Error: channels.json.example not found!"
    echo "   Make sure you're running this script from the GoondVR directory."
    exit 1
fi
