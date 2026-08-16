export interface FileNameParts {
  prefix: string;
  suffix: string;
  extension: string;
}

export class FileNameUtils {
  static extractFirstH1(markdown: string): string | null {
    let inFence = false;
    let fenceMarker = '';

    for (const line of markdown.split(/\r?\n/)) {
      const fenceMatch = line.match(/^\s*(```+|~~~+)/);
      if (fenceMatch) {
        const marker = fenceMatch[1][0];
        if (!inFence) {
          inFence = true;
          fenceMarker = marker;
        } else if (marker === fenceMarker) {
          inFence = false;
          fenceMarker = '';
        }
        continue;
      }

      if (inFence) {
        continue;
      }

      const h1Match = line.match(/^#\s+(.+?)\s*#*\s*$/);
      if (h1Match) {
        return h1Match[1].trim();
      }
    }

    return null;
  }

  static slugifyTitle(title: string | null | undefined): string {
    if (!title) {
      return '';
    }

    return title
      .trim()
      .toLowerCase()
      .replace(/[\\/]+/g, '-')
      .replace(/\s+/g, '-')
      .replace(/[^\p{Letter}\p{Number}-]+/gu, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '');
  }

  static isValidSuffix(suffix: string): boolean {
    const trimmedSuffix = suffix.trim();
    return !!trimmedSuffix
      && trimmedSuffix !== '.'
      && trimmedSuffix !== '..'
      && !trimmedSuffix.includes('/')
      && !trimmedSuffix.includes('\\')
      && !trimmedSuffix.includes('..');
  }

  static splitFileName(fileName: string, generatorType?: string): FileNameParts {
    const extensionMatch = fileName.match(/(\.[^.]*)$/);
    const extension = extensionMatch?.[1] || '.md';
    const baseName = extensionMatch ? fileName.slice(0, -extension.length) : fileName;
    let prefix = '';

    if (!generatorType || generatorType === 'date') {
      prefix = baseName.match(/^\d{4}-\d{2}-\d{2}-/)?.[0] || '';
    }

    if (!prefix && (!generatorType || generatorType === 'sequence')) {
      prefix = baseName.match(/^\d+-/)?.[0] || '';
    }

    return {
      prefix,
      suffix: baseName.slice(prefix.length),
      extension
    };
  }

  static buildFileName(prefix: string, suffix: string, extension = '.md'): string {
    return `${prefix}${suffix}${extension || '.md'}`;
  }
}
