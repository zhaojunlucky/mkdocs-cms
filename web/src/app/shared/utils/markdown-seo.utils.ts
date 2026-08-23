import { fromMarkdown } from 'mdast-util-from-markdown';
import { toString } from 'mdast-util-to-string';

export class MarkdownSeoUtils {
  static extractFirstH1(markdown: string): string {
    const heading = this.children(markdown).find(node => node.type === 'heading' && node.depth === 1);
    return heading ? this.normalizeText(toString(heading)) : '';
  }

  static countH1(markdown: string): number {
    return this.children(markdown).filter(node => node.type === 'heading' && node.depth === 1).length;
  }

  static generateDescription(markdown: string, maxLength: number): string {
    const nodes = this.children(markdown);
    const excerptIndex = nodes.findIndex(node => this.isMaterialExcerptMarker(node));
    const beforeExcerpt = excerptIndex >= 0 ? nodes.slice(0, excerptIndex) : [];
    const beforeExcerptText = this.firstMeaningfulParagraph(beforeExcerpt);
    if (beforeExcerptText) {
      return this.truncateAtWord(beforeExcerptText, maxLength);
    }

    const firstH1Index = nodes.findIndex(node => node.type === 'heading' && node.depth === 1);
    const afterH1 = firstH1Index >= 0 ? nodes.slice(firstH1Index + 1) : nodes;
    const afterH1Text = this.firstMeaningfulParagraph(afterH1);
    if (afterH1Text) {
      return this.truncateAtWord(afterH1Text, maxLength);
    }

    const fallbackText = this.firstMeaningfulParagraph(nodes);
    return fallbackText ? this.truncateAtWord(fallbackText, maxLength) : '';
  }

  static isSimilarTitle(title: string, h1: string): boolean {
    const normalizedTitle = this.normalizeComparableText(title);
    const normalizedH1 = this.normalizeComparableText(h1);
    return normalizedTitle === normalizedH1
      || normalizedTitle.includes(normalizedH1)
      || normalizedH1.includes(normalizedTitle);
  }

  private static children(markdown: string): any[] {
    return (fromMarkdown(markdown || '') as any).children || [];
  }

  private static firstMeaningfulParagraph(nodes: any[]): string {
    for (const node of nodes) {
      if (node.type !== 'paragraph') {
        continue;
      }

      const text = this.normalizeText(toString(node));
      if (this.isMeaningfulDescription(text)) {
        return text;
      }
    }

    return '';
  }

  private static isMaterialExcerptMarker(node: any): boolean {
    return node.type === 'html' && /^<!--\s*more\s*-->$/i.test(String(node.value || '').trim());
  }

  private static isMeaningfulDescription(text: string): boolean {
    const cjkCharacters = text.match(/[\p{Script=Han}\p{Script=Hiragana}\p{Script=Katakana}\p{Script=Hangul}]/gu)?.length || 0;
    if (cjkCharacters >= 12) {
      return true;
    }

    if (text.length < 40) {
      return false;
    }

    const words = text.match(/[A-Za-z0-9][A-Za-z0-9'-]*/g) || [];
    if (words.length < 8) {
      return false;
    }

    return /[\p{Letter}]{3,}/u.test(text);
  }

  private static normalizeText(text: string): string {
    return text
      .replace(/<!--[\s\S]*?-->/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
  }

  private static truncateAtWord(value: string, maxLength: number): string {
    if (value.length <= maxLength) {
      return value;
    }

    const truncated = value.slice(0, maxLength + 1);
    const lastSpace = truncated.lastIndexOf(' ');
    const boundary = lastSpace > 60 ? lastSpace : maxLength;
    return value.slice(0, boundary).trim().replace(/[,.!?;:]$/, '');
  }

  private static normalizeComparableText(value: string): string {
    return value.toLowerCase().replace(/[^a-z0-9]+/g, ' ').trim();
  }
}
