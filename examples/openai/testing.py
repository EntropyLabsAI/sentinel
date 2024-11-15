from duckduckgo_search import DDGS
from typing import List, Dict, Optional
import requests
from bs4 import BeautifulSoup

def search_duckduckgo_and_get_content(
    query: str,
    max_results: int = 5,
    region: str = "wt-wt",
    safesearch: str = "moderate"
) -> List[Dict]:
    """
    Search DuckDuckGo, return results, and fetch content from the first few links.
    
    Args:
        query (str): Search query
        max_results (int): Maximum number of results to return (default: 5)
        region (str): Region code (default: "wt-wt" for worldwide)
        safesearch (str): SafeSearch setting ("on", "moderate", or "off")
    
    Returns:
        List[Dict]: List of search results, each containing 'title', 'href', 'snippet', and 'content'
    """
    try:
        with DDGS() as ddgs:
            results = []
            search_results = ddgs.text(
                query,
                region=region,
                safesearch=safesearch,
                max_results=max_results
            )
            
            for r in search_results:
                result = {
                    'title': r['title'],
                    'href': r['href'],
                    'snippet': r['body'],
                    'content': ''
                }
                try:
                    response = requests.get(r['href'], timeout=5)
                    soup = BeautifulSoup(response.content, 'html.parser')
                    # Extract text content from the web page
                    paragraphs = soup.find_all('p')
                    text_content = ' '.join([p.get_text() for p in paragraphs])
                    result['content'] = text_content
                except Exception as e:
                    print(f"Error fetching content from {r['href']}: {str(e)}")
                results.append(result)
                
            return results
        
    except Exception as e:
        print(f"Error performing search: {str(e)}")
        return []

# Example usage
if __name__ == "__main__":
    custom_results = search_duckduckgo_and_get_content(
        query="events in san francisco this weekend",
        max_results=3,
        region="us-en",
        safesearch="on"
    )
    
    # Print results
    for i, result in enumerate(custom_results):
        print(f"\nResult {i}:")
        print(f"Title: {result['title']}")
        print(f"URL: {result['href']}")
        print(f"Snippet: {result['snippet']}\n")
        print(f"Content:\n{result['content'][:500]}")  # Print first 500 characters of content
