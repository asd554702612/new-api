import fs from "node:fs/promises";
import path from "node:path";
import { SpreadsheetFile, Workbook } from "@oai/artifact-tool";

const inputPath = "/tmp/new-api-export/new_api_export.json";
const outputDir = "/Users/mac/work/gepin/new-api/outputs/user_subscription_export";
const outputPath = path.join(outputDir, "users_subscriptions.xlsx");
const previewDir = path.join(outputDir, "previews");

const raw = JSON.parse(await fs.readFile(inputPath, "utf8"));
const now = new Date();

const users = raw.users ?? [];
const plans = raw.subscription_plans ?? [];
const subs = raw.user_subscriptions ?? [];
const orders = raw.subscription_orders ?? [];

const userById = new Map(users.map((u) => [num(u.id), u]));
const planById = new Map(plans.map((p) => [num(p.id), p]));
const latestOrderByUserPlan = new Map(Object.entries(raw.latest_order_by_user_plan ?? {}));

const workbook = Workbook.create();
const overview = workbook.worksheets.add("概览");
const userSheet = workbook.worksheets.add("用户明细");
const subSheet = workbook.worksheets.add("订阅明细");
const checkSheet = workbook.worksheets.add("校验结果");

writeUserSheet(userSheet);
writeSubscriptionSheet(subSheet);
writeCheckSheet(checkSheet);
writeOverview(overview);

for (const sheet of [overview, userSheet, subSheet, checkSheet]) {
  sheet.showGridLines = false;
}

const inspectSummary = await workbook.inspect({
  kind: "workbook,sheet,table",
  maxChars: 6000,
  tableMaxRows: 4,
  tableMaxCols: 8,
});
console.log("INSPECT_SUMMARY");
console.log(inspectSummary.ndjson);

const formulaErrors = await workbook.inspect({
  kind: "match",
  searchTerm: "#REF!|#DIV/0!|#VALUE!|#NAME\\?|#N/A",
  options: { useRegex: true, maxResults: 300 },
  summary: "final formula error scan",
});
console.log("FORMULA_ERRORS");
console.log(formulaErrors.ndjson);

await fs.mkdir(previewDir, { recursive: true });
const previewRanges = {
  "概览": "A1:D15",
  "用户明细": "A1:AD25",
  "订阅明细": "A1:AH30",
  "校验结果": "A1:J30",
};
for (const [sheetName, range] of Object.entries(previewRanges)) {
  const preview = await workbook.render({
    sheetName,
    range,
    scale: 1,
    format: "png",
  });
  await fs.writeFile(
    path.join(previewDir, `${sheetName}.png`),
    new Uint8Array(await preview.arrayBuffer()),
  );
}

await fs.mkdir(outputDir, { recursive: true });
const xlsx = await SpreadsheetFile.exportXlsx(workbook);
await xlsx.save(outputPath);
console.log(`SAVED ${outputPath}`);

function writeOverview(sheet) {
  const activeCount = subs.filter((s) => s.status === "active").length;
  const expiredCount = subs.filter((s) => s.status === "expired").length;
  const cancelledCount = subs.filter((s) => s.status === "cancelled").length;
  const checkRows = buildValidationRows();
  const needsReviewCount = checkRows.filter((r) => r.finalStatus !== "通过").length;
  const activeExpiredMismatch = subs.filter((s) => s.status === "active" && unixDate(s.end_time) && unixDate(s.end_time) <= now).length;
  const overUsedCount = subs.filter((s) => num(s.amount_total) > 0 && num(s.amount_used) > num(s.amount_total)).length;

  const rows = [
    ["new-api 用户与订阅导出", "", "", ""],
    ["导出时间", raw.exported_at ?? "", "数据库", raw.db_name ?? ""],
    ["数据源", "82.156.54.174 / Docker 容器 new-api / PostgreSQL", "敏感字段", "已排除密码、access token、2FA 密钥"],
    ["", "", "", ""],
    ["指标", "数量", "说明", "校验口径"],
    ["用户总数", users.length, "users 表导出行数", "包含未删除与软删除用户的非敏感字段"],
    ["套餐总数", plans.length, "subscription_plans 表导出行数", "当前套餐配置"],
    ["订阅总数", subs.length, "user_subscriptions 表导出行数", "包含 active/expired/cancelled 等状态"],
    ["订阅订单总数", orders.length, "subscription_orders 表导出行数", "用于金额参考"],
    ["有效订阅", activeCount, "status = active", ""],
    ["过期订阅", expiredCount, "status = expired", ""],
    ["取消订阅", cancelledCount, "status = cancelled", ""],
    ["需检查订阅", needsReviewCount, "校验结果不为通过", "详见“校验结果”工作表"],
    ["状态过期不一致", activeExpiredMismatch, "active 但结束时间早于当前日期", ""],
    ["用量超额", overUsedCount, "amount_total > 0 且 amount_used > amount_total", ""],
  ];

  writeMatrix(sheet, rows);
  sheet.getRange("A1:D1").merge();
  sheet.getRange("A1:D1").format = {
    fill: "#1F4E79",
    font: { bold: true, color: "#FFFFFF", size: 16 },
    horizontalAlignment: "center",
  };
  sheet.getRange("A5:D5").format = headerFormat();
  sheet.getRange("A2:D3").format = {
    fill: "#EAF3F8",
    borders: { preset: "outside", style: "thin", color: "#B7C9D6" },
  };
  sheet.getRange("A6:D15").format = {
    borders: { preset: "inside", style: "thin", color: "#D9E2EA" },
  };
  sheet.getRange("B6:B15").format.numberFormat = "#,##0";
  setWidths(sheet, [22, 14, 42, 52], rows.length);
}

function writeUserSheet(sheet) {
  const headers = [
    "用户ID", "用户名", "显示名", "邮箱", "手机号", "角色", "角色值", "状态", "状态值", "用户分组",
    "余额额度", "已用额度", "请求次数", "邀请码", "邀请人数", "邀请剩余额度", "邀请历史额度", "邀请人ID",
    "GitHub ID", "Discord ID", "OIDC ID", "微信 ID", "Telegram ID", "LinuxDO ID", "备注", "Stripe Customer",
    "注册时间", "最后登录时间", "是否删除", "删除时间",
  ];
  const rows = users.map((u) => [
    num(u.id), text(u.username), text(u.display_name), text(u.email), text(u.phone_number),
    roleLabel(u.role), num(u.role), userStatusLabel(u.status), num(u.status), text(u.group),
    num(u.quota), num(u.used_quota), num(u.request_count), text(u.aff_code), num(u.aff_count),
    num(u.aff_quota), num(u.aff_history), emptyToNull(u.inviter_id),
    text(u.github_id), text(u.discord_id), text(u.oidc_id), text(u.wechat_id), text(u.telegram_id), text(u.linux_do_id),
    text(u.remark), text(u.stripe_customer), unixDate(u.created_at), unixDate(u.last_login_at),
    u.deleted_at ? "是" : "否", parseDateOrNull(u.deleted_at),
  ]);
  writeTableSheet(sheet, "UsersTable", headers, rows, [27, 28, 30]);
  setWidths(sheet, [10, 18, 18, 28, 18, 12, 10, 12, 10, 16, 14, 14, 12, 12, 12, 14, 14, 12, 22, 22, 22, 22, 22, 22, 28, 22, 20, 20, 10, 20], rows.length + 1);
  setNumberFormat(sheet, rows.length + 1, [11, 12, 13, 15, 16, 17], "#,##0");
}

function writeSubscriptionSheet(sheet) {
  const headers = [
    "订阅ID", "用户ID", "用户名", "邮箱", "手机号", "套餐ID", "套餐标题", "套餐副标题", "套餐启用",
    "套餐价格", "币种", "周期", "重置周期", "订阅状态", "来源", "总额度", "已用额度", "剩余额度",
    "使用率", "开始时间", "结束时间", "距离过期天数", "上次重置", "下次重置", "升级分组", "原分组",
    "最新订单ID", "最新订单金额", "最新支付方式", "最新订单状态", "最新订单完成时间", "金额校验",
    "完整性校验", "最终结论",
  ];
  const rows = subs.map((s) => {
    const user = userById.get(num(s.user_id)) ?? {};
    const plan = planById.get(num(s.plan_id)) ?? {};
    const order = latestOrderByUserPlan.get(`${num(s.user_id)}:${num(s.plan_id)}`) ?? {};
    return [
      num(s.id), num(s.user_id), text(user.username), text(user.email), text(user.phone_number),
      num(s.plan_id), text(plan.title), text(plan.subtitle), boolText(plan.enabled),
      parseMoney(plan.price_amount), text(plan.currency), durationText(plan), resetText(plan),
      text(s.status), text(s.source), num(s.amount_total), num(s.amount_used), null, null,
      unixDate(s.start_time), unixDate(s.end_time), null, unixDate(s.last_reset_time), unixDate(s.next_reset_time),
      text(s.upgrade_group), text(s.prev_user_group), emptyToNull(order.id), parseMoney(order.money),
      text(order.payment_method), text(order.status), unixDate(order.complete_time), null, null, null,
    ];
  });
  writeTableSheet(sheet, "SubscriptionsTable", headers, rows, [20, 21, 23, 24, 31]);
  if (rows.length > 0) {
    const end = rows.length + 1;
    sheet.getRange(`R2`).formulas = [["=IF(P2=0,\"无限\",MAX(P2-Q2,0))"]];
    sheet.getRange(`R2:R${end}`).fillDown();
    sheet.getRange(`S2`).formulas = [["=IF(P2=0,\"无限\",IF(P2=\"\",\"\",Q2/P2))"]];
    sheet.getRange(`S2:S${end}`).fillDown();
    sheet.getRange(`V2`).formulas = [["=IF(U2=\"\",\"\",INT(U2-TODAY()))"]];
    sheet.getRange(`V2:V${end}`).fillDown();
    sheet.getRange(`AF2`).formulas = [["=IF(AA2=\"\",\"无订单\",IF(AND(J2<>\"\",AB2<>\"\",ROUND(J2,2)<>ROUND(AB2,2)),\"金额需确认\",\"金额匹配/无需确认\"))"]];
    sheet.getRange(`AF2:AF${end}`).fillDown();
    sheet.getRange(`AG2`).formulas = [["=IF(OR(B2=\"\",F2=\"\",G2=\"\",T2=\"\",U2=\"\"),\"信息不完整\",IF(AND(P2>0,Q2>P2),\"用量超额\",IF(AND(N2=\"active\",U2<=TODAY()),\"已过期但状态未更新\",\"通过\")))"]];
    sheet.getRange(`AG2:AG${end}`).fillDown();
    sheet.getRange(`AH2`).formulas = [["=IF(OR(AF2=\"金额需确认\",AG2<>\"通过\"),\"需检查\",\"通过\")"]];
    sheet.getRange(`AH2:AH${end}`).fillDown();
    setNumberFormat(sheet, end, [10, 28], "#,##0.00");
    setNumberFormat(sheet, end, [16, 17, 18], "#,##0");
    setNumberFormat(sheet, end, [19], "0.0%");
    setNumberFormat(sheet, end, [22], "#,##0");
  }
  setWidths(sheet, [10, 10, 18, 28, 18, 10, 22, 36, 12, 12, 10, 16, 14, 12, 12, 14, 14, 14, 12, 20, 20, 14, 20, 20, 16, 16, 12, 14, 16, 14, 20, 20, 20, 14], rows.length + 1);
}

function writeCheckSheet(sheet) {
  const headers = ["订阅ID", "用户ID", "用户名", "套餐ID", "套餐标题", "订阅状态", "金额校验", "完整性校验", "最终结论", "问题说明"];
  const rows = buildValidationRows().map((r) => [
    r.subscriptionId, r.userId, r.username, r.planId, r.planTitle, r.subStatus,
    r.amountStatus, r.integrityStatus, r.finalStatus, r.note,
  ]);
  writeTableSheet(sheet, "ValidationTable", headers, rows, []);
  setWidths(sheet, [10, 10, 18, 10, 24, 12, 18, 22, 12, 70], rows.length + 1);
}

function buildValidationRows() {
  return subs.map((s) => {
    const user = userById.get(num(s.user_id));
    const plan = planById.get(num(s.plan_id));
    const order = latestOrderByUserPlan.get(`${num(s.user_id)}:${num(s.plan_id)}`);
    const issues = [];
    if (!user) issues.push("用户不存在");
    if (!plan) issues.push("套餐不存在");
    if (!plan?.title) issues.push("套餐标题缺失");
    if (!unixDate(s.start_time)) issues.push("开始时间缺失");
    if (!unixDate(s.end_time)) issues.push("结束时间缺失");
    if (num(s.amount_total) > 0 && num(s.amount_used) > num(s.amount_total)) issues.push("已用额度超过总额度");
    if (s.status === "active" && unixDate(s.end_time) && unixDate(s.end_time) <= now) issues.push("active 订阅已到期但状态未更新");

    let amountStatus = "无订单";
    if (order) {
      const planPrice = parseMoney(plan?.price_amount);
      const orderMoney = parseMoney(order.money);
      amountStatus = Number.isFinite(planPrice) && Number.isFinite(orderMoney) && Math.round(planPrice * 100) !== Math.round(orderMoney * 100)
        ? "金额需确认"
        : "金额匹配/无需确认";
      if (amountStatus === "金额需确认") issues.push("最新订单金额与当前套餐价格不同");
    }

    const integrityStatus = issues.length === 0 || (issues.length === 1 && issues[0].startsWith("最新订单金额"))
      ? "通过"
      : issues.filter((i) => !i.startsWith("最新订单金额")).join("；") || "通过";
    const finalStatus = amountStatus === "金额需确认" || integrityStatus !== "通过" ? "需检查" : "通过";
    return {
      subscriptionId: num(s.id),
      userId: num(s.user_id),
      username: text(user?.username),
      planId: num(s.plan_id),
      planTitle: text(plan?.title),
      subStatus: text(s.status),
      amountStatus,
      integrityStatus,
      finalStatus,
      note: issues.join("；") || "未发现异常",
    };
  });
}

function writeTableSheet(sheet, tableName, headers, rows, dateColumns) {
  const matrix = [headers, ...rows];
  writeMatrix(sheet, matrix);
  sheet.freezePanes.freezeRows(1);
  sheet.getRangeByIndexes(0, 0, 1, headers.length).format = headerFormat();
  if (matrix.length > 1) {
    const tableRange = `A1:${colName(headers.length)}${matrix.length}`;
    const table = sheet.tables.add(tableRange, true, tableName);
    table.style = "TableStyleMedium2";
  }
  for (const col of dateColumns) {
    setNumberFormat(sheet, matrix.length, [col], "yyyy-mm-dd hh:mm");
  }
}

function writeMatrix(sheet, matrix) {
  if (!matrix.length || !matrix[0].length) return;
  sheet.getRangeByIndexes(0, 0, matrix.length, matrix[0].length).values = matrix;
}

function setWidths(sheet, widths, rowCount) {
  widths.forEach((width, idx) => {
    sheet.getRangeByIndexes(0, idx, Math.max(rowCount, 1), 1).format.columnWidth = width;
  });
}

function setNumberFormat(sheet, rowCount, oneBasedColumns, format) {
  for (const col of oneBasedColumns) {
    sheet.getRangeByIndexes(1, col - 1, Math.max(rowCount - 1, 1), 1).format.numberFormat = format;
  }
}

function headerFormat() {
  return {
    fill: "#1F4E79",
    font: { bold: true, color: "#FFFFFF" },
    horizontalAlignment: "center",
    verticalAlignment: "center",
    wrapText: true,
    borders: { preset: "outside", style: "thin", color: "#173A5E" },
  };
}

function colName(n) {
  let s = "";
  while (n > 0) {
    const mod = (n - 1) % 26;
    s = String.fromCharCode(65 + mod) + s;
    n = Math.floor((n - mod) / 26);
  }
  return s;
}

function num(v) {
  if (v === null || v === undefined || v === "") return 0;
  const n = Number(v);
  return Number.isFinite(n) ? n : 0;
}

function parseMoney(v) {
  if (v === null || v === undefined || v === "") return null;
  const n = Number(v);
  return Number.isFinite(n) ? n : null;
}

function text(v) {
  if (v === null || v === undefined) return "";
  return String(v);
}

function emptyToNull(v) {
  if (v === null || v === undefined || v === "" || Number(v) === 0) return null;
  return num(v);
}

function boolText(v) {
  if (v === true) return "是";
  if (v === false) return "否";
  return "";
}

function unixDate(v) {
  const n = num(v);
  if (!n) return null;
  return new Date(n * 1000);
}

function parseDateOrNull(v) {
  if (!v) return null;
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? null : d;
}

function roleLabel(v) {
  switch (num(v)) {
    case 1:
      return "普通用户";
    case 10:
      return "管理员";
    case 100:
      return "根用户";
    default:
      return "未知";
  }
}

function userStatusLabel(v) {
  switch (num(v)) {
    case 1:
      return "启用";
    case 2:
      return "禁用";
    default:
      return "未知";
  }
}

function durationText(plan) {
  if (!plan) return "";
  if (plan.duration_unit === "custom") return `${num(plan.custom_seconds)} 秒`;
  const value = num(plan.duration_value);
  const unitMap = { year: "年", month: "月", day: "天", hour: "小时" };
  return `${value}${unitMap[plan.duration_unit] ?? plan.duration_unit ?? ""}`;
}

function resetText(plan) {
  if (!plan) return "";
  const period = text(plan.quota_reset_period);
  if (period === "custom") return `${num(plan.quota_reset_custom_seconds)} 秒`;
  const map = { never: "不重置", daily: "每日", weekly: "每周", monthly: "每月" };
  return map[period] ?? period;
}
