#!/usr/bin/env node

// Bundle size analysis script for the frontend
const fs = require('fs');
const path = require('path');

// Get the build directory
const buildDir = path.join(__dirname, '..', 'frontend', 'dist');
const assetsDir = path.join(buildDir, 'assets');

// Function to format bytes to human readable format
function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Function to analyze bundle sizes
async function analyzeBundleSizes() {
  try {
    // Check if build directory exists
    if (!fs.existsSync(buildDir)) {
      console.log('Build directory not found. Please run "npm run build" first.');
      return;
    }

    // Get all files in the assets directory
    const files = fs.readdirSync(assetsDir);
    
    // Separate JS, CSS, and other files
    const jsFiles = files.filter(file => file.endsWith('.js'));
    const cssFiles = files.filter(file => file.endsWith('.css'));
    const otherFiles = files.filter(file => !file.endsWith('.js') && !file.endsWith('.css'));
    
    console.log('=== Bundle Size Analysis ===\n');
    
    // Analyze JS files
    console.log('JavaScript Files:');
    let totalJsSize = 0;
    jsFiles.forEach(file => {
      const filePath = path.join(assetsDir, file);
      const stats = fs.statSync(filePath);
      const size = stats.size;
      totalJsSize += size;
      console.log(`  ${file}: ${formatBytes(size)}`);
    });
    console.log(`  Total JS Size: ${formatBytes(totalJsSize)}\n`);
    
    // Analyze CSS files
    console.log('CSS Files:');
    let totalCssSize = 0;
    cssFiles.forEach(file => {
      const filePath = path.join(assetsDir, file);
      const stats = fs.statSync(filePath);
      const size = stats.size;
      totalCssSize += size;
      console.log(`  ${file}: ${formatBytes(size)}`);
    });
    console.log(`  Total CSS Size: ${formatBytes(totalCssSize)}\n`);
    
    // Analyze other files
    console.log('Other Files:');
    let totalOtherSize = 0;
    otherFiles.forEach(file => {
      const filePath = path.join(assetsDir, file);
      const stats = fs.statSync(filePath);
      const size = stats.size;
      totalOtherSize += size;
      console.log(`  ${file}: ${formatBytes(size)}`);
    });
    console.log(`  Total Other Size: ${formatBytes(totalOtherSize)}\n`);
    
    // Total size
    const totalSize = totalJsSize + totalCssSize + totalOtherSize;
    console.log(`Total Bundle Size: ${formatBytes(totalSize)}`);
    
    // Optimization suggestions
    console.log('\n=== Optimization Suggestions ===');
    if (totalJsSize > 500000) { // 500KB
      console.log('⚠️  JavaScript bundle is quite large. Consider:');
      console.log('   - Code splitting for route-based chunks');
      console.log('   - Tree shaking to remove unused code');
      console.log('   - Lazy loading non-critical components');
    }
    
    if (totalCssSize > 100000) { // 100KB
      console.log('⚠️  CSS bundle is quite large. Consider:');
      console.log('   - Removing unused CSS rules');
      console.log('   - Using CSS modules for better scoping');
      console.log('   - Minifying CSS in production');
    }
    
    // Check for large individual files
    const largeFiles = files
      .map(file => {
        const filePath = path.join(assetsDir, file);
        const stats = fs.statSync(filePath);
        return { name: file, size: stats.size };
      })
      .filter(file => file.size > 100000) // 100KB
      .sort((a, b) => b.size - a.size);
    
    if (largeFiles.length > 0) {
      console.log('\n⚠️  Large individual files (>100KB):');
      largeFiles.forEach(file => {
        console.log(`   ${file.name}: ${formatBytes(file.size)}`);
      });
    }
    
  } catch (error) {
    console.error('Error analyzing bundle sizes:', error.message);
  }
}

// Run the analysis
analyzeBundleSizes();