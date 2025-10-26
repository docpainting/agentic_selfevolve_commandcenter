# YouTube Retrieval with Transcript Extraction

## Overview

YouTube is a **critical source** for bleeding-edge concepts, tutorials, and expert explanations. Many developers, researchers, and thought leaders share cutting-edge knowledge through video before it appears in written documentation. The agent needs to search YouTube and extract transcripts for analysis.

## Why YouTube is Critical

### 1. **Bleeding-Edge Content**
```
Written documentation: 6-12 months behind
Conference talks on YouTube: Real-time
    ↓
Example: New Go 1.23 features explained by core team
         Available on YouTube weeks before official docs
```

### 2. **Expert Explanations**
```
Stack Overflow: Community answers
YouTube: Direct from experts (Go team, library authors, etc.)
    ↓
Example: "How to optimize Go handlers" by Google engineer
         More authoritative than random blog post
```

### 3. **Visual Demonstrations**
```
Written tutorial: "Add caching to your handler"
YouTube tutorial: Shows exact code, debugging, results
    ↓
Agent can extract transcript + understand visual context
```

### 4. **Conference Talks & Workshops**
```
GopherCon, KubeCon, etc. → All on YouTube
Latest research, best practices, case studies
    ↓
Often 6-12 months ahead of written documentation
```

## YouTube Search Strategy

### Search Query Generation

```go
type YouTubeQuery struct {
    Query string
    Filters YouTubeFilters
    MaxResults int
}

type YouTubeFilters struct {
    UploadDate string  // "today", "week", "month", "year", "all"
    Duration string    // "short" (<4min), "medium" (4-20min), "long" (>20min)
    Features []string  // "subtitles", "hd", "4k"
    SortBy string      // "relevance", "upload_date", "view_count", "rating"
}

func generateYouTubeQuery(context *Context, confidence ConfidenceScore) YouTubeQuery {
    // Base query from task context
    baseQuery := fmt.Sprintf("%s %s tutorial", context.TaskType, context.Language)
    
    // Add specificity based on confidence gaps
    if confidence.Factors["past_experience"] == 0.0 {
        baseQuery += " beginner guide"
    }
    if confidence.Factors["pattern_availability"] == 0.0 {
        baseQuery += " best practices"
    }
    
    // Add recency for bleeding-edge topics
    uploadDate := "year"  // Default: past year
    if isCuttingEdge(context) {
        uploadDate = "month"  // Bleeding-edge: past month
    }
    
    return YouTubeQuery{
        Query: baseQuery,
        Filters: YouTubeFilters{
            UploadDate: uploadDate,
            Duration: "medium",  // 4-20 min (sweet spot)
            Features: []string{"subtitles"},  // Must have transcripts
            SortBy: "relevance",
        },
        MaxResults: 10,
    }
}

func isCuttingEdge(context *Context) bool {
    cuttingEdgeKeywords := []string{
        "new", "latest", "2025", "2024",
        "experimental", "preview", "beta",
        "cutting-edge", "state-of-the-art",
    }
    
    for _, keyword := range cuttingEdgeKeywords {
        if strings.Contains(strings.ToLower(context.Description), keyword) {
            return true
        }
    }
    
    return false
}
```

### YouTube API Integration

```go
type YouTubeSearcher struct {
    apiKey string
    client *http.Client
}

type YouTubeVideo struct {
    ID string
    Title string
    ChannelName string
    Description string
    PublishedAt time.Time
    Duration time.Duration
    ViewCount int64
    LikeCount int64
    URL string
    ThumbnailURL string
}

func (ys *YouTubeSearcher) Search(query YouTubeQuery) ([]YouTubeVideo, error) {
    // Build YouTube Data API v3 request
    params := url.Values{}
    params.Set("part", "snippet")
    params.Set("q", query.Query)
    params.Set("type", "video")
    params.Set("maxResults", strconv.Itoa(query.MaxResults))
    params.Set("key", ys.apiKey)
    
    // Apply filters
    if query.Filters.UploadDate != "all" {
        publishedAfter := getPublishedAfter(query.Filters.UploadDate)
        params.Set("publishedAfter", publishedAfter.Format(time.RFC3339))
    }
    
    if query.Filters.Duration != "" {
        params.Set("videoDuration", query.Filters.Duration)
    }
    
    if contains(query.Filters.Features, "subtitles") {
        params.Set("videoCaption", "closedCaption")
    }
    
    params.Set("order", query.Filters.SortBy)
    
    // Make API request
    apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?%s", params.Encode())
    resp, err := ys.client.Get(apiURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Parse response
    var result YouTubeSearchResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    // Convert to YouTubeVideo objects
    videos := []YouTubeVideo{}
    for _, item := range result.Items {
        video := YouTubeVideo{
            ID: item.ID.VideoID,
            Title: item.Snippet.Title,
            ChannelName: item.Snippet.ChannelTitle,
            Description: item.Snippet.Description,
            PublishedAt: item.Snippet.PublishedAt,
            URL: fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID.VideoID),
            ThumbnailURL: item.Snippet.Thumbnails.High.URL,
        }
        videos = append(videos, video)
    }
    
    // Get additional details (duration, view count, etc.)
    videos = ys.enrichVideoDetails(videos)
    
    return videos, nil
}
```

## Transcript Extraction

### Multiple Extraction Methods

```go
type TranscriptExtractor interface {
    Extract(videoID string) (*Transcript, error)
}

// Method 1: YouTube API (official captions)
type YouTubeAPIExtractor struct {
    apiKey string
}

// Method 2: youtube-transcript-api (Python library via subprocess)
type PythonTranscriptExtractor struct {
    pythonPath string
}

// Method 3: yt-dlp (fallback for when API fails)
type YTDLPExtractor struct {
    ytdlpPath string
}

type Transcript struct {
    VideoID string
    Language string
    Segments []TranscriptSegment
    FullText string
    Duration time.Duration
}

type TranscriptSegment struct {
    Text string
    Start time.Duration
    End time.Duration
}
```

### Method 1: YouTube API (Official Captions)

```go
func (ye *YouTubeAPIExtractor) Extract(videoID string) (*Transcript, error) {
    // 1. Get available caption tracks
    captionsURL := fmt.Sprintf(
        "https://www.googleapis.com/youtube/v3/captions?videoId=%s&part=snippet&key=%s",
        videoID, ye.apiKey,
    )
    
    resp, err := http.Get(captionsURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var captionsResult CaptionsResult
    json.NewDecoder(resp.Body).Decode(&captionsResult)
    
    // 2. Find English caption track
    var captionID string
    for _, item := range captionsResult.Items {
        if item.Snippet.Language == "en" {
            captionID = item.ID
            break
        }
    }
    
    if captionID == "" {
        return nil, fmt.Errorf("no English captions available")
    }
    
    // 3. Download caption track
    downloadURL := fmt.Sprintf(
        "https://www.googleapis.com/youtube/v3/captions/%s?key=%s",
        captionID, ye.apiKey,
    )
    
    resp, err = http.Get(downloadURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 4. Parse caption format (usually SRT or VTT)
    transcript := parseCaption(resp.Body)
    transcript.VideoID = videoID
    
    return transcript, nil
}
```

### Method 2: youtube-transcript-api (Python Library)

```go
func (pe *PythonTranscriptExtractor) Extract(videoID string) (*Transcript, error) {
    // Use youtube-transcript-api Python library
    script := fmt.Sprintf(`
import json
from youtube_transcript_api import YouTubeTranscriptApi

try:
    transcript = YouTubeTranscriptApi.get_transcript('%s')
    print(json.dumps(transcript))
except Exception as e:
    print(json.dumps({"error": str(e)}))
`, videoID)
    
    cmd := exec.Command(pe.pythonPath, "-c", script)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    // Parse JSON output
    var segments []map[string]interface{}
    if err := json.Unmarshal(output, &segments); err != nil {
        return nil, err
    }
    
    // Convert to Transcript
    transcript := &Transcript{
        VideoID: videoID,
        Language: "en",
        Segments: []TranscriptSegment{},
    }
    
    fullText := ""
    for _, seg := range segments {
        segment := TranscriptSegment{
            Text: seg["text"].(string),
            Start: time.Duration(seg["start"].(float64) * float64(time.Second)),
        }
        transcript.Segments = append(transcript.Segments, segment)
        fullText += segment.Text + " "
    }
    
    transcript.FullText = strings.TrimSpace(fullText)
    
    return transcript, nil
}
```

### Method 3: yt-dlp (Fallback)

```go
func (ye *YTDLPExtractor) Extract(videoID string) (*Transcript, error) {
    // Use yt-dlp to download subtitles
    videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
    
    cmd := exec.Command(ye.ytdlpPath,
        "--skip-download",
        "--write-auto-sub",
        "--sub-lang", "en",
        "--sub-format", "vtt",
        "-o", "/tmp/%(id)s.%(ext)s",
        videoURL,
    )
    
    if err := cmd.Run(); err != nil {
        return nil, err
    }
    
    // Read downloaded subtitle file
    subtitlePath := fmt.Sprintf("/tmp/%s.en.vtt", videoID)
    content, err := os.ReadFile(subtitlePath)
    if err != nil {
        return nil, err
    }
    
    // Parse VTT format
    transcript := parseVTT(content)
    transcript.VideoID = videoID
    
    // Clean up
    os.Remove(subtitlePath)
    
    return transcript, nil
}
```

### Unified Transcript Extraction with Fallbacks

```go
type TranscriptService struct {
    extractors []TranscriptExtractor
}

func NewTranscriptService(apiKey, pythonPath, ytdlpPath string) *TranscriptService {
    return &TranscriptService{
        extractors: []TranscriptExtractor{
            &PythonTranscriptExtractor{pythonPath: pythonPath},  // Try Python library first (fastest)
            &YouTubeAPIExtractor{apiKey: apiKey},                // Then official API
            &YTDLPExtractor{ytdlpPath: ytdlpPath},              // Finally yt-dlp (slowest but most reliable)
        },
    }
}

func (ts *TranscriptService) Extract(videoID string) (*Transcript, error) {
    var lastErr error
    
    for i, extractor := range ts.extractors {
        log.Info("Attempting transcript extraction", 
                 "method", i+1, 
                 "extractor", fmt.Sprintf("%T", extractor))
        
        transcript, err := extractor.Extract(videoID)
        if err == nil {
            log.Info("Transcript extracted successfully", 
                     "method", i+1,
                     "segments", len(transcript.Segments))
            return transcript, nil
        }
        
        log.Warn("Transcript extraction failed, trying next method",
                 "method", i+1,
                 "error", err)
        lastErr = err
    }
    
    return nil, fmt.Errorf("all extraction methods failed: %w", lastErr)
}
```

## Transcript Analysis

### Extract Key Information from Transcript

```go
type TranscriptAnalysis struct {
    Summary string
    KeyPoints []string
    CodeExamples []CodeExample
    Timestamps []Timestamp
    Concepts []string
    Recommendations []string
}

type Timestamp struct {
    Time time.Duration
    Description string
    Relevance float64
}

func analyzeTranscript(transcript *Transcript, context *Context) (*TranscriptAnalysis, error) {
    // Use LLM to analyze transcript
    prompt := fmt.Sprintf(`
Analyze this YouTube video transcript for information relevant to: %s

Transcript:
%s

Extract:
1. Summary (2-3 sentences)
2. Key points (bullet points)
3. Code examples mentioned (with timestamps)
4. Important concepts explained
5. Recommendations or best practices
6. Most relevant timestamps for the task

Provide structured output.
`, context.Goal, transcript.FullText)
    
    response := llm.Generate(prompt)
    
    analysis := parseTranscriptAnalysis(response)
    
    // Add video metadata
    analysis.VideoID = transcript.VideoID
    analysis.Duration = transcript.Duration
    
    return analysis, nil
}
```

### Semantic Search Within Transcript

```go
func searchTranscript(transcript *Transcript, query string) []TranscriptSegment {
    // Use embeddings to find most relevant segments
    queryEmbedding := embeddings.Embed(query)
    
    relevantSegments := []TranscriptSegment{}
    
    for _, segment := range transcript.Segments {
        segmentEmbedding := embeddings.Embed(segment.Text)
        similarity := cosineSimilarity(queryEmbedding, segmentEmbedding)
        
        if similarity > 0.7 {
            segment.Relevance = similarity
            relevantSegments = append(relevantSegments, segment)
        }
    }
    
    // Sort by relevance
    sort.Slice(relevantSegments, func(i, j int) bool {
        return relevantSegments[i].Relevance > relevantSegments[j].Relevance
    })
    
    return relevantSegments
}
```

## Enhanced Retrieval with YouTube

### Updated Retrieval Sources

```go
type RetrievalSources struct {
    Web bool
    GitHub bool
    StackOverflow bool
    Docs bool
    ArXiv bool
    YouTube bool  // NEW!
}

func retrieveInformation(query RetrievalQuery) (*RetrievalResult, error) {
    results := &RetrievalResult{
        Query: query.Query,
        Documents: []Document{},
        YouTubeVideos: []YouTubeDocument{},  // NEW!
    }
    
    var wg sync.WaitGroup
    
    // ... existing web, GitHub, Stack Overflow, docs searches ...
    
    // NEW: YouTube search
    if contains(query.Sources, "youtube") {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            // Search YouTube
            videos, err := youtubeSearcher.Search(generateYouTubeQuery(query))
            if err != nil {
                log.Warn("YouTube search failed", "error", err)
                return
            }
            
            // Extract transcripts for top videos
            for i, video := range videos {
                if i >= 5 {  // Limit to top 5 videos
                    break
                }
                
                transcript, err := transcriptService.Extract(video.ID)
                if err != nil {
                    log.Warn("Transcript extraction failed", 
                             "video_id", video.ID, 
                             "error", err)
                    continue
                }
                
                // Analyze transcript
                analysis, err := analyzeTranscript(transcript, query.Context)
                if err != nil {
                    log.Warn("Transcript analysis failed", "error", err)
                    continue
                }
                
                // Create YouTube document
                doc := YouTubeDocument{
                    Video: video,
                    Transcript: transcript,
                    Analysis: analysis,
                    Relevance: calculateRelevance(analysis, query),
                }
                
                results.YouTubeVideos = append(results.YouTubeVideos, doc)
            }
        }()
    }
    
    wg.Wait()
    
    // Rank all results (including YouTube)
    results.RankAllSources()
    
    // Synthesize insights (including YouTube)
    results.Synthesis = synthesizeWithYouTube(results)
    
    return results, nil
}
```

### YouTube Document Structure

```go
type YouTubeDocument struct {
    Video YouTubeVideo
    Transcript *Transcript
    Analysis *TranscriptAnalysis
    Relevance float64
    Source string  // "youtube"
}

func (yd *YouTubeDocument) GetSummary() string {
    return fmt.Sprintf(`
Video: %s
Channel: %s
Published: %s
Duration: %s
Views: %d

Summary: %s

Key Points:
%s

Most Relevant Timestamps:
%s
`, 
        yd.Video.Title,
        yd.Video.ChannelName,
        yd.Video.PublishedAt.Format("2006-01-02"),
        yd.Video.Duration,
        yd.Video.ViewCount,
        yd.Analysis.Summary,
        strings.Join(yd.Analysis.KeyPoints, "\n- "),
        formatTimestamps(yd.Analysis.Timestamps),
    )
}
```

## Enhanced Synthesis with YouTube

```go
func synthesizeWithYouTube(results *RetrievalResult) *Synthesis {
    synthesis := &Synthesis{
        KeyInsights: []string{},
        BestPractices: []string{},
        CommonMistakes: []string{},
        CodeExamples: []CodeExample{},
        RelevantPatterns: []Pattern{},
        YouTubeRecommendations: []YouTubeRecommendation{},  // NEW!
    }
    
    // Extract from YouTube videos
    for _, ytDoc := range results.YouTubeVideos {
        // Add key points
        synthesis.KeyInsights = append(synthesis.KeyInsights, ytDoc.Analysis.KeyPoints...)
        
        // Add code examples with video source
        for _, example := range ytDoc.Analysis.CodeExamples {
            example.Source = fmt.Sprintf("YouTube: %s (at %s)", 
                ytDoc.Video.Title, 
                example.Timestamp)
            synthesis.CodeExamples = append(synthesis.CodeExamples, example)
        }
        
        // Add recommendations
        synthesis.BestPractices = append(synthesis.BestPractices, ytDoc.Analysis.Recommendations...)
        
        // Create YouTube recommendation
        rec := YouTubeRecommendation{
            Video: ytDoc.Video,
            Relevance: ytDoc.Relevance,
            KeyTimestamps: ytDoc.Analysis.Timestamps,
            Summary: ytDoc.Analysis.Summary,
        }
        synthesis.YouTubeRecommendations = append(synthesis.YouTubeRecommendations, rec)
    }
    
    // ... existing synthesis from web, GitHub, etc. ...
    
    // Deduplicate and rank
    synthesis.Deduplicate()
    synthesis.RankByRelevance()
    
    return synthesis
}
```

## Example: Bleeding-Edge Retrieval

### User Request
```
User: "Implement the new Go 1.23 iterator pattern in my code"
```

### Confidence Check
```
Confidence: 0.2 (very low!)
  - Past experience: 0.0 (never used Go 1.23 iterators)
  - Pattern availability: 0.0 (no patterns in Neo4j)
  - Code understanding: 0.1 (new language feature)
  - Strategy clarity: 0.0 (unclear how to implement)
  - Risk assessment: 0.1 (new feature, potential bugs)
```

### Retrieval Triggered
```
Query: "Go 1.23 iterator pattern tutorial"
Sources: web, github, stackoverflow, docs, youtube
Upload date filter: "month" (bleeding-edge!)
```

### YouTube Results
```
1. "Go 1.23 Iterators Explained" by Go Team
   - Published: 2 weeks ago
   - Duration: 15:23
   - Views: 45,000
   - Transcript extracted ✓
   
   Key Points:
   - New range-over-func pattern
   - How to implement custom iterators
   - Performance considerations
   - Common mistakes to avoid
   
   Most Relevant Timestamps:
   - 2:15 - Basic iterator implementation
   - 5:30 - Custom iterator example
   - 10:45 - Performance comparison
   - 13:20 - Error handling patterns

2. "GopherCon 2024: Iterator Patterns" by Core Team Member
   - Published: 1 month ago
   - Duration: 28:45
   - Views: 120,000
   - Transcript extracted ✓
   
   Key Points:
   - Design rationale behind iterators
   - Real-world use cases
   - Migration from old patterns
   - Best practices from production use
   
   Code Examples:
   - Timestamp 8:30: Basic iterator
   - Timestamp 15:20: Error handling
   - Timestamp 22:10: Performance optimization

3. "Go 1.23 New Features Deep Dive" by Popular YouTuber
   - Published: 3 weeks ago
   - Duration: 42:15
   - Views: 85,000
   - Transcript extracted ✓
```

### Synthesis
```
Key Insights:
- Go 1.23 introduces range-over-func pattern for iterators
- Replaces old channel-based iteration (more efficient)
- Requires func(yield func(T) bool) signature
- Zero allocation in most cases

Best Practices (from videos):
- Use yield function to emit values
- Return early if yield returns false (consumer stopped)
- Handle errors via separate error channel or return value
- Test with fuzz testing for edge cases

Code Examples:
From "Go 1.23 Iterators Explained" (2:15):
```go
func Count(n int) func(yield func(int) bool) {
    return func(yield func(int) bool) {
        for i := 0; i < n; i++ {
            if !yield(i) {
                return
            }
        }
    }
}

// Usage
for v := range Count(10) {
    fmt.Println(v)
}
```

Common Mistakes (from GopherCon talk):
- Forgetting to check yield return value
- Not handling early termination
- Overusing iterators where simple loops suffice

YouTube Recommendations:
1. Watch "Go 1.23 Iterators Explained" (15:23)
   - Timestamps: 2:15, 5:30, 10:45
   - Relevance: 0.95
   
2. Watch "GopherCon 2024: Iterator Patterns" (28:45)
   - Timestamps: 8:30, 15:20, 22:10
   - Relevance: 0.92
```

### Confidence Boost
```
Old confidence: 0.2
    ↓
Retrieved YouTube transcripts with:
- Official Go team explanation
- Conference talk from core team
- Multiple code examples
- Best practices from production use
    ↓
New confidence: 0.9 (high!)
    ↓
Proceed with implementation
```

## Configuration

```yaml
# config/youtube_retrieval.yaml
youtube:
  enabled: true
  
  api_key: ${YOUTUBE_API_KEY}
  
  search:
    max_results: 10
    default_upload_date: "year"
    bleeding_edge_upload_date: "month"
    preferred_duration: "medium"  # 4-20 minutes
    require_subtitles: true
  
  transcript:
    extraction_methods:
      - python_library  # youtube-transcript-api (fastest)
      - youtube_api     # Official API
      - ytdlp          # yt-dlp (most reliable)
    
    max_videos_to_process: 5
    timeout_per_video: 30  # seconds
  
  analysis:
    extract_key_points: true
    extract_code_examples: true
    extract_timestamps: true
    extract_concepts: true
    
    llm_model: "gemma3:27b"
  
  caching:
    enabled: true
    ttl: 86400  # Cache transcripts for 24 hours
  
  rate_limiting:
    max_requests_per_day: 10000  # YouTube API quota
```

## Installation Requirements

```bash
# Install youtube-transcript-api (Python)
pip3 install youtube-transcript-api

# Install yt-dlp (fallback)
pip3 install yt-dlp

# Or via system package manager
sudo apt install yt-dlp
```

## Benefits

### 1. **Bleeding-Edge Knowledge**
Access to latest features, techniques, and best practices before written docs

### 2. **Expert Explanations**
Direct from language creators, library authors, and industry experts

### 3. **Visual Context**
Understand not just what to do, but how (with demonstrations)

### 4. **Conference Talks**
Access to GopherCon, KubeCon, etc. with cutting-edge research

### 5. **Multiple Perspectives**
Different experts explain same concept in different ways

### 6. **Timestamped Learning**
Jump directly to relevant sections via timestamp extraction

## Summary

### YouTube Integration
- ✅ Search YouTube with smart query generation
- ✅ Filter by upload date (bleeding-edge: past month)
- ✅ Extract transcripts (3 methods with fallbacks)
- ✅ Analyze transcripts with LLM
- ✅ Extract key points, code examples, timestamps
- ✅ Semantic search within transcripts
- ✅ Synthesize with other sources

### Confidence Boost
- ✅ YouTube provides information not available elsewhere
- ✅ Especially critical for new/bleeding-edge topics
- ✅ Can boost confidence from 0.2 → 0.9

### Use Cases
- ✅ New language features (Go 1.23 iterators)
- ✅ Cutting-edge frameworks (before docs exist)
- ✅ Conference talks (latest research)
- ✅ Expert tutorials (from authoritative sources)
- ✅ Visual demonstrations (complex concepts)

The agent now has access to **bleeding-edge knowledge from YouTube** with full transcript extraction and analysis! 🎥

