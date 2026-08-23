import { StrUtils } from './str.utils';

describe('StrUtils', () => {
  it('parses redirect edit URLs without losing encoded path separators', () => {
    const result = StrUtils.parseRedirectUrl('/repositories/1/redirect/edit?path=blog%2FPosts%2F2026%2Ffile.md');

    expect(result['paths']).toEqual(['/repositories', '1', 'redirect', 'edit']);
    expect(result['queryParams']).toEqual({
      path: 'blog/Posts/2026/file.md'
    });
  });

  it('parses redirect URLs without query params', () => {
    const result = StrUtils.parseRedirectUrl('/home');

    expect(result['paths']).toEqual(['/home']);
    expect(result['queryParams']).toEqual({});
  });
});
