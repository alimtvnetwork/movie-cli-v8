# State
- Fixed boolean flag naming for UseTable -> IsTableOutput, UseJson -> IsJsonOutput, and UserProvidedPath -> IsUserProvidedPath across the cmd package.
- Extracted OutputFormatOpts and embedded it into ScanLoopConfig and ScanOutputOpts in cmd/types.go.
- Fixed struct literals in movie_scan.go and movie_scan_loop.go.
- Wrapped arguments in GetRecommendations in tmdb/client.go to fix line lengths > 100 characters.
- Ran go vet and go fmt.
- Recorded modified files using 33-test-inventory-generator.py.
