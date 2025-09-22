import pandas as pd
import sys
import os

def csv_to_excel(csv_file, excel_file):
    """Convert CSV file to Excel format"""
    try:
        # Check if input file exists
        if not os.path.exists(csv_file):
            print(f"Error: Input file {csv_file} does not exist")
            return False
            
        # Read CSV file
        df = pd.read_csv(csv_file)
        
        # Write to Excel file
        df.to_excel(excel_file, index=False)
        
        print(f"Successfully converted {csv_file} to {excel_file}")
        return True
    except Exception as e:
        print(f"Error converting file: {e}")
        return False

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python csv_to_excel.py <input.csv> <output.xlsx>")
        sys.exit(1)
    
    csv_file = sys.argv[1]
    excel_file = sys.argv[2]
    
    success = csv_to_excel(csv_file, excel_file)
    if not success:
        sys.exit(1)