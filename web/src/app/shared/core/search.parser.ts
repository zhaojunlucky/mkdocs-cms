/**
 * Search parser for advanced search syntax
 */
export class SearchParser {
  private _content: string = '';
  private _filters: Map<string, boolean | string> = new Map();

  constructor(searchTerm: string) {
    this.parse(searchTerm);
  }

  /**
   * Parse the search term to extract content and filters
   */
  private parse(searchTerm: string): void {
    if (!searchTerm) {
      return;
    }

    const term = searchTerm.toLowerCase().trim();

    // Extract filters
    // Draft filter
    if (term.includes('draft:')) {
      const draftMatch = term.match(/draft:(true|false)/);
      if (draftMatch) {
        this._filters.set('draft', draftMatch[1] === 'true');
      }
    }

    // Extract content (everything that's not a filter)
    this._content = term
      .replace(/draft:(true|false)/g, '')
      .trim();
  }

  /**
   * Get the content part of the search (non-filter text)
   */
  get content(): string {
    return this._content;
  }

  /**
   * Get all filters as a Map
   */
  get filters(): Map<string, boolean | string> {
    return this._filters;
  }

  /**
   * Check if a specific filter exists
   */
  hasFilter(key: string): boolean {
    return this._filters.has(key);
  }

  /**
   * Get the value of a specific filter
   */
  getFilter(key: string): boolean | string | undefined {
    return this._filters.get(key);
  }
}
