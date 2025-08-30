#!/bin/bash

# 批量替换admin.go中的响应方法
file="internal/handlers/admin.go"

# BadRequest responses
sed -i '' 's/c\.JSON(http\.StatusBadRequest, gin\.H{$/common.BadRequestResponse(c, /g' "$file"
sed -i '' 's/^\s*"code":\s*400,$/"""/g' "$file"
sed -i '' 's/^\s*"message":\s*"\([^"]*\)",*$/\1)/g' "$file"

# InternalServerError responses  
sed -i '' 's/c\.JSON(http\.StatusInternalServerError, gin\.H{$/common.InternalServerErrorResponse(c, /g' "$file"
sed -i '' 's/^\s*"code":\s*500,$/"""/g' "$file"

# NotFound responses
sed -i '' 's/c\.JSON(http\.StatusNotFound, gin\.H{$/common.NotFoundResponse(c, /g' "$file"
sed -i '' 's/^\s*"code":\s*404,$/"""/g' "$file"

# Success responses
sed -i '' 's/c\.JSON(http\.StatusOK, gin\.H{$/common.SuccessResponse(c, gin.H{/g' "$file"
sed -i '' 's/^\s*"code":\s*200,$/"""/g' "$file"
sed -i '' 's/^\s*"message":\s*"success",$/"""/g' "$file"

echo "批量替换完成"
