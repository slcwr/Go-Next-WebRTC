import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // 認証が必要なパス
  const protectedPaths = ['/todos', '/profile'];

  // 認証済みユーザーがアクセスできないパス
  const authPaths = ['/login', '/register'];

  const isProtectedPath = protectedPaths.some((path) => pathname.startsWith(path));
  const isAuthPath = authPaths.some((path) => pathname.startsWith(path));

  // クッキーまたはローカルストレージの代わりに、
  // Next.jsのミドルウェアではヘッダーをチェック
  // ここではシンプルに実装（実際の認証チェックはクライアント側で行う）

  if (isProtectedPath) {
    // クライアント側で認証チェックを行うため、ここでは通過させる
    return NextResponse.next();
  }

  if (isAuthPath) {
    // 既にログイン済みの場合はリダイレクト（クライアント側で処理）
    return NextResponse.next();
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * 以下を除くすべてのパスにマッチ:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};
