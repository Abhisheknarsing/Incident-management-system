import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function UploadPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Upload Incident Data</h1>
        <p className="text-muted-foreground">
          Upload Excel files containing incident data for analysis
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>File Upload</CardTitle>
          <CardDescription>
            Select an Excel file (.xlsx, .xls) containing incident data
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="border-2 border-dashed border-muted-foreground/25 rounded-lg p-8 text-center">
            <p className="text-muted-foreground">Upload functionality will be implemented in the next task</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}