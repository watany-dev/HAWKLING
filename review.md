# パフォーマンス改善レビュー

## 現状の問題点

1. `main.go` における不必要なフィルタリングの重複
2. AWS APIリクエストの非効率な使用
3. スライスの初期化方法が非効率
4. フォーマッタにおけるメモリ使用の最適化不足
5. フィルタリングフラグ（--used, --unused, --days）が正しく機能していない

## 改善提案

### 1. フィルタリングロジックの最適化 (main.go)

現在、`listRoles`関数では複数のフィルタ条件（days, onlyUsed, onlyUnused）がそれぞれ独立して適用されており、複数回のスライス走査が発生しています。これを1回の走査で済むように最適化できます。

```go
// 改善前: 複数回のフィルタリングループ
if days > 0 { /* フィルタリング処理 */ }
if onlyUsed { /* フィルタリング処理 */ }
if onlyUnused { /* フィルタリング処理 */ }

// 改善後: 1回のループで全条件を評価
var filteredRoles []aws.Role
for _, role := range roles {
    if (days > 0 && !role.IsUnused(days)) ||
       (onlyUsed && role.LastUsed == nil) ||
       (onlyUnused && role.LastUsed != nil) {
        continue
    }
    filteredRoles = append(filteredRoles, role)
}
```

### 2. フィルタリングフラグの修正 (list.go)

現在の実装では、フィルタリングロジックに問題があります：

```go
// 問題点: フィルタリングロジックが逆になっている
if (c.options.OnlyUsed && !isUnused) || (c.options.OnlyUnused && isUnused) {
    filteredRoles = append(filteredRoles, role)
}
```

この実装では：
1. `OnlyUsed`フラグが立っていると、`!isUnused`（つまり使用中）のロールをフィルタリングしていますが、`isUnused`が誤った判定をしています。
2. `OnlyUnused`フラグが立っていると、`isUnused`（未使用）のロールをフィルタリングしていますが、日数の考慮が不十分です。

修正案：
```go
// 修正案: フィルタリングロジックの明確化
for _, role := range roles {
    // days > 0 の条件を適切に扱う
    isUnusedForDays := role.IsUnused(c.options.Days)
    
    // OnlyUsed: LastUsed != nil（使用されている）のロールだけを表示
    if c.options.OnlyUsed && (role.LastUsed == nil) {
        continue
    }
    
    // OnlyUnused: 指定日数間使用されていないロールだけを表示
    if c.options.OnlyUnused && !isUnusedForDays {
        continue
    }
    
    filteredRoles = append(filteredRoles, role)
}
```

### 3. AWS APIリクエストの最適化 (awsclient.go)

`ListRoles`メソッドでは現在、ロールの取得とロールの最終使用日時の取得が分離されています。AWS APIが提供する情報を最大限活用することで、APIコールを減らせる可能性があります。

また、ゴルーチンの起動方法を改善することでオーバーヘッドを減らせます。

```go
// 改善前: ゴルーチン起動のオーバーヘッド
for i := range roles {
    go func(i int) {
        // ...
    }(i)
}

// 改善後: ワーカープールパターンの採用
```

### 4. IsUnusedロジックの修正 (iam.go)

現在の`IsUnused`メソッドのロジックも確認が必要です：

```go
// IsUnused checks if a role is unused for the specified number of days
func (r *Role) IsUnused(days int) bool {
    if r.LastUsed == nil {
        return true  // LastUsed が nil なら未使用
    }

    threshold := time.Now().AddDate(0, 0, -days)
    return r.LastUsed.Before(threshold)  // days日以前に使われていれば未使用と判断
}
```

このロジックでは、`LastUsed`が`nil`（一度も使われていない）の場合は常に「未使用」として扱われます。また閾値との比較は、最後に使われた日が閾値より前（古い）場合に「未使用」と判断しています。これは論理的に正しいです。

### 5. スライス初期化の最適化

多くの場所でスライスが事前容量指定なしで初期化されています。容量を事前に指定することで、動的な拡張によるメモリ割り当てを減らせます。

```go
// 改善前
var unusedRoles []aws.Role

// 改善後
unusedRoles := make([]aws.Role, 0, len(roles))
```

### 6. formatter.goのメモリ使用最適化

`FormatRolesAsJSON`関数では、すべてのロール情報がJSONとしてメモリに格納されてから出力されます。大量のロールがある場合、ストリーミング方式の出力に変更することで、メモリ使用量を削減できます。

## 実装優先度

1. **最高**: フィルタリングフラグの修正（機能面の問題解決）
   - ✅ 実装完了: フィルタリングロジックを修正し、--used, --unused, --daysフラグが正しく機能するように修正
   - ✅ テスト追加: フィルタリングロジックのテストを追加
2. **高**: スライス初期化の最適化（実装が容易かつ効果が明確）
   - ✅ 部分的に実装: filteredRoles スライスの初期化を最適化
3. **中**: フィルタリングロジックの統合（コード可読性を維持しつつ改善）
   - ✅ 実装完了: フィルタリングロジックを1回のループで済むように最適化
4. **中**: ゴルーチン管理の最適化（現状でもセマフォがあり基本的には問題ない）
5. **低**: JSONフォーマッタのストリーミング出力（大量データ時のみ効果あり）

これらの改善により、メモリ使用量の削減とCPU使用効率の向上が期待できます。特に大量のロールを処理する場合に効果が顕著になるでしょう。