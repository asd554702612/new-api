/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useMemo, useState } from 'react';
import {
  Badge,
  Button,
  Card,
  Divider,
  Select,
  Skeleton,
  Space,
  Tag,
  Tooltip,
  Typography,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess, renderQuota } from '../../helpers';
import { RefreshCw, Sparkles } from 'lucide-react';
import SubscriptionPurchaseModal from './modals/SubscriptionPurchaseModal';
import OfficialPaymentModal from './modals/OfficialPaymentModal';
import {
  formatSubscriptionDuration,
  formatSubscriptionPrice,
  formatSubscriptionResetPeriod,
} from '../../helpers/subscriptionFormat';
import {
  getDefaultOfficialTradeType,
  isCasdoorPayment,
  isOfficialPaymentMethod,
  isSafeOfficialCheckoutUrl,
  normalizeOfficialPaymentResult,
} from '../../helpers/officialPayment';

const { Text } = Typography;

const saleBlockReasonText = {
  disabled: '套餐未启用',
  before_start: '套餐尚未开始售卖',
  after_end: '套餐已下架',
  daily_window: '当前不在每日售卖时间内',
  weekly_day: '今日不可购买该套餐',
  sold_out: '今日已售罄',
  purchase_once: '有效期内只能购买一次该套餐',
};

const weekdayText = {
  1: '周一',
  2: '周二',
  3: '周三',
  4: '周四',
  5: '周五',
  6: '周六',
  7: '周日',
};

function planSaleAvailability(planDTO) {
  return planDTO?.sale_availability || null;
}

function planCanBePurchased(planDTO) {
  const availability = planSaleAvailability(planDTO);
  return !availability || availability.available !== false;
}

function planUnavailableText(planDTO, t) {
  const availability = planSaleAvailability(planDTO);
  if (!availability || availability.available !== false) return '';
  return (
    availability.block_message ||
    t(saleBlockReasonText[availability.block_reason] || '套餐当前不可购买')
  );
}

function weeklySaleDaysLabel(days = [], t) {
  if (!Array.isArray(days) || days.length === 0) return '';
  return days.map((day) => t(weekdayText[day] || String(day))).join('/');
}

function formatCountdown(seconds, t) {
  const value = Number(seconds || 0);
  if (value <= 0) return '';
  const hours = Math.floor(value / 3600);
  const minutes = Math.floor((value % 3600) / 60);
  if (hours > 0) return `${hours}${t('小时')}${minutes}${t('分钟')}`;
  return `${minutes}${t('分钟')}`;
}

// 过滤易支付方式
function getEpayMethods(payMethods = []) {
  return (payMethods || []).filter(
    (m) =>
      m?.type &&
      m.type !== 'stripe' &&
      m.type !== 'creem' &&
      m.type !== 'waffo_pancake' &&
      !isOfficialPaymentMethod(m.type),
  );
}

// 提交易支付表单
function submitEpayForm({ url, params }) {
  const form = document.createElement('form');
  form.action = url;
  form.method = 'POST';
  const isSafari =
    navigator.userAgent.indexOf('Safari') > -1 &&
    navigator.userAgent.indexOf('Chrome') < 1;
  if (!isSafari) form.target = '_blank';
  Object.keys(params || {}).forEach((key) => {
    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = key;
    input.value = params[key];
    form.appendChild(input);
  });
  document.body.appendChild(form);
  form.submit();
  document.body.removeChild(form);
}

const SubscriptionPlansCard = ({
  t,
  loading = false,
  plans = [],
  userQuota = 0,
  payMethods = [],
  enableOnlineTopUp = false,
  enableStripeTopUp = false,
  enableCreemTopUp = false,
  enableWechatPayTopUp = false,
  enableAlipayTopUp = false,
  enableCasdoorTopUp = false,
  billingPreference,
  onChangeBillingPreference,
  activeSubscriptions = [],
  allSubscriptions = [],
  reloadSubscriptionSelf,
  reloadUserQuota,
  withCard = true,
}) => {
  const [open, setOpen] = useState(false);
  const [selectedPlan, setSelectedPlan] = useState(null);
  const [paying, setPaying] = useState(false);
  const [selectedEpayMethod, setSelectedEpayMethod] = useState('');
  const [refreshing, setRefreshing] = useState(false);
  const [officialPaymentOpen, setOfficialPaymentOpen] = useState(false);
  const [officialPayment, setOfficialPayment] = useState(null);

  const epayMethods = useMemo(() => getEpayMethods(payMethods), [payMethods]);

  const openBuy = (p) => {
    if (!planCanBePurchased(p)) {
      showError(planUnavailableText(p, t));
      return;
    }
    setSelectedPlan(p);
    setSelectedEpayMethod(epayMethods?.[0]?.type || '');
    setOpen(true);
  };

  const closeBuy = () => {
    setOpen(false);
    setSelectedPlan(null);
    setPaying(false);
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await reloadSubscriptionSelf?.();
    } finally {
      setRefreshing(false);
    }
  };

  const payStripe = async () => {
    if (!selectedPlan?.plan?.stripe_price_id) {
      showError(t('该套餐未配置 Stripe'));
      return;
    }
    setPaying(true);
    try {
      const res = await API.post('/api/subscription/stripe/pay', {
        plan_id: selectedPlan.plan.id,
      });
      if (res.data?.message === 'success') {
        window.open(res.data.data?.pay_link, '_blank');
        showSuccess(t('已打开支付页面'));
        closeBuy();
      } else {
        const errorMsg =
          typeof res.data?.data === 'string'
            ? res.data.data
            : res.data?.message || t('支付失败');
        showError(errorMsg);
      }
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  const payCreem = async () => {
    if (!selectedPlan?.plan?.creem_product_id) {
      showError(t('该套餐未配置 Creem'));
      return;
    }
    setPaying(true);
    try {
      const res = await API.post('/api/subscription/creem/pay', {
        plan_id: selectedPlan.plan.id,
      });
      if (res.data?.message === 'success') {
        window.open(res.data.data?.checkout_url, '_blank');
        showSuccess(t('已打开支付页面'));
        closeBuy();
      } else {
        const errorMsg =
          typeof res.data?.data === 'string'
            ? res.data.data
            : res.data?.message || t('支付失败');
        showError(errorMsg);
      }
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  const payBalance = async () => {
    if (!selectedPlan?.plan?.id) {
      showError(t('请选择订阅套餐'));
      return;
    }
    setPaying(true);
    try {
      const res = await API.post('/api/subscription/balance/pay', {
        plan_id: selectedPlan.plan.id,
      });
      if (res.data?.message === 'success') {
        showSuccess(t('购买成功'));
        closeBuy();
        await Promise.all([reloadSubscriptionSelf?.(), reloadUserQuota?.()]);
      } else {
        const errorMsg =
          typeof res.data?.data === 'string'
            ? res.data.data
            : res.data?.message || t('支付失败');
        showError(errorMsg);
      }
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  const payEpay = async () => {
    if (!selectedEpayMethod) {
      showError(t('请选择支付方式'));
      return;
    }
    setPaying(true);
    try {
      const res = await API.post('/api/subscription/epay/pay', {
        plan_id: selectedPlan.plan.id,
        payment_method: selectedEpayMethod,
      });
      if (res.data?.message === 'success') {
        submitEpayForm({ url: res.data.url, params: res.data.data });
        showSuccess(t('已发起支付'));
        closeBuy();
      } else {
        const errorMsg =
          typeof res.data?.data === 'string'
            ? res.data.data
            : res.data?.message || t('支付失败');
        showError(errorMsg);
      }
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  const planSupportsOfficialPay = (plan) => {
    const currency = String(plan?.currency || 'USD').toUpperCase();
    return currency === '' || currency === 'USD' || currency === 'CNY';
  };

  const handleOfficialPaymentResponse = async (res) => {
    if (!res) {
      showError(t('支付请求失败'));
      return;
    }
    const { success, message, data } = res.data || {};
    if (success !== true && message !== 'success') {
      const errorMsg =
        typeof data === 'string' ? data : message || t('支付失败');
      showError(errorMsg);
      return;
    }

    const result = normalizeOfficialPaymentResult(data || {});
    if (result.kind === 'redirect') {
      if (isSafeOfficialCheckoutUrl(result.url)) {
        window.location.href = result.url;
        closeBuy();
      } else {
        showError(t('支付跳转地址不安全'));
      }
      return;
    }

    if (result.kind === 'qr') {
      setOfficialPayment({ ...result, tradeNo: data?.trade_no || '' });
      setOfficialPaymentOpen(true);
      closeBuy();
      return;
    }

    if (result.kind === 'jsapi') {
      if (
        typeof window === 'undefined' ||
        !window.WeixinJSBridge ||
        !result.jsapiParams
      ) {
        showError(t('当前环境无法拉起微信 JSAPI 支付，请换用 H5 或 Native 支付'));
        return;
      }
      window.WeixinJSBridge.invoke(
        'getBrandWCPayRequest',
        result.jsapiParams,
        (response) => {
          if (response?.err_msg === 'get_brand_wcpay_request:ok') {
            showSuccess(t('已发起支付，请等待到账'));
            closeBuy();
          } else if (
            response?.err_msg !== 'get_brand_wcpay_request:cancel'
          ) {
            showError(t('微信 JSAPI 支付拉起失败'));
          }
        },
      );
      return;
    }

    showError(t('支付响应缺少跳转地址或二维码'));
  };

  useEffect(() => {
    const tradeNo = officialPayment?.tradeNo;
    if (!officialPaymentOpen || !tradeNo) {
      return undefined;
    }

    let stopped = false;
    let attempts = 0;
    const maxAttempts = 40;

    const pollOrderStatus = async () => {
      if (stopped) {
        return;
      }
      attempts += 1;
      try {
        const res = await API.get(
          `/api/payment/orders/my?keyword=${encodeURIComponent(
            tradeNo,
          )}&page=1&page_size=1`,
        );
        const items = res?.data?.data?.items || [];
        const order =
          items.find((item) => item.trade_no === tradeNo) || items[0];
        const status = String(order?.status || '').toUpperCase();
        if (status === 'COMPLETED') {
          stopped = true;
          setOfficialPaymentOpen(false);
          setOfficialPayment(null);
          showSuccess(t('更新成功'));
          reloadSubscriptionSelf?.();
          return;
        }
        if (
          ['FAILED', 'EXPIRED', 'CANCELLED', 'REFUNDED'].includes(status)
        ) {
          stopped = true;
          setOfficialPaymentOpen(false);
          setOfficialPayment(null);
          showError(t('支付未完成或已关闭'));
        }
      } catch (error) {
        // Keep waiting; transient network errors should not close the QR modal.
      }
      if (attempts >= maxAttempts) {
        stopped = true;
      }
    };

    pollOrderStatus();
    const timer = window.setInterval(pollOrderStatus, 3000);
    return () => {
      stopped = true;
      window.clearInterval(timer);
    };
  }, [officialPaymentOpen, officialPayment?.tradeNo, reloadSubscriptionSelf, t]);

  const payOfficial = async (paymentType) => {
    if (!selectedPlan?.plan?.id) {
      showError(t('请选择订阅套餐'));
      return;
    }
    if (!planSupportsOfficialPay(selectedPlan.plan)) {
      showError(t('官方微信/支付宝暂不支持该套餐币种'));
      return;
    }

    setPaying(true);
    try {
      const endpoint =
        paymentType === 'wechat_pay'
          ? '/api/subscription/wechat-pay/pay'
          : isCasdoorPayment(paymentType)
            ? '/api/subscription/casdoor/pay'
            : '/api/subscription/alipay/pay';
      const payload = {
        plan_id: selectedPlan.plan.id,
      };
      const tradeType = getDefaultOfficialTradeType(paymentType);
      if (tradeType) {
        payload.trade_type = tradeType;
      }
      const res = await API.post(endpoint, payload);
      await handleOfficialPaymentResponse(res);
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  // 当前订阅信息 - 支持多个订阅
  const hasActiveSubscription = activeSubscriptions.length > 0;
  const hasAnySubscription = allSubscriptions.length > 0;
  const disableSubscriptionPreference = !hasActiveSubscription;
  const isSubscriptionPreference =
    billingPreference === 'subscription_first' ||
    billingPreference === 'subscription_only';
  const displayBillingPreference =
    disableSubscriptionPreference && isSubscriptionPreference
      ? 'wallet_first'
      : billingPreference;
  const subscriptionPreferenceLabel =
    billingPreference === 'subscription_only' ? t('仅用订阅') : t('优先订阅');

  const planPurchaseCountMap = useMemo(() => {
    const map = new Map();
    (allSubscriptions || []).forEach((sub) => {
      const planId = sub?.subscription?.plan_id;
      if (!planId) return;
      map.set(planId, (map.get(planId) || 0) + 1);
    });
    return map;
  }, [allSubscriptions]);

  const planTitleMap = useMemo(() => {
    const map = new Map();
    (plans || []).forEach((p) => {
      const plan = p?.plan;
      if (!plan?.id) return;
      map.set(plan.id, plan.title || '');
    });
    return map;
  }, [plans]);

  const getPlanPurchaseCount = (planId) =>
    planPurchaseCountMap.get(planId) || 0;

  // 计算单个订阅的剩余天数
  const getRemainingDays = (sub) => {
    if (!sub?.subscription?.end_time) return 0;
    const now = Date.now() / 1000;
    const remaining = sub.subscription.end_time - now;
    return Math.max(0, Math.ceil(remaining / 86400));
  };

  // 计算单个订阅的使用进度
  const getUsagePercent = (sub) => {
    const total = Number(sub?.subscription?.amount_total || 0);
    const used = Number(sub?.subscription?.amount_used || 0);
    if (total <= 0) return 0;
    return Math.round((used / total) * 100);
  };

  const cardContent = (
    <>
      {/* 卡片头部 */}
      {loading ? (
        <div className='space-y-4'>
          {/* 我的订阅骨架屏 */}
          <Card className='!rounded-xl w-full' bodyStyle={{ padding: '12px' }}>
            <div className='flex items-center justify-between mb-3'>
              <Skeleton.Title active style={{ width: 100, height: 20 }} />
              <Skeleton.Button active style={{ width: 24, height: 24 }} />
            </div>
            <div className='space-y-2'>
              <Skeleton.Paragraph active rows={2} />
            </div>
          </Card>
          {/* 套餐列表骨架屏 */}
          <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-5 w-full px-1'>
            {[1, 2, 3].map((i) => (
              <Card
                key={i}
                className='!rounded-xl w-full h-full'
                bodyStyle={{ padding: 16 }}
              >
                <Skeleton.Title
                  active
                  style={{ width: '60%', height: 24, marginBottom: 8 }}
                />
                <Skeleton.Paragraph
                  active
                  rows={1}
                  style={{ marginBottom: 12 }}
                />
                <div className='text-center py-4'>
                  <Skeleton.Title
                    active
                    style={{ width: '40%', height: 32, margin: '0 auto' }}
                  />
                </div>
                <Skeleton.Paragraph active rows={3} style={{ marginTop: 12 }} />
                <Skeleton.Button
                  active
                  block
                  style={{ marginTop: 16, height: 32 }}
                />
              </Card>
            ))}
          </div>
        </div>
      ) : (
        <Space vertical style={{ width: '100%' }} spacing={8}>
          {/* 当前订阅状态 */}
          <Card className='!rounded-xl w-full' bodyStyle={{ padding: '12px' }}>
            <div className='flex items-center justify-between mb-2 gap-3'>
              <div className='flex items-center gap-2 flex-1 min-w-0'>
                <Text strong>{t('我的订阅')}</Text>
                {hasActiveSubscription ? (
                  <Tag
                    color='white'
                    size='small'
                    shape='circle'
                    prefixIcon={<Badge dot type='success' />}
                  >
                    {activeSubscriptions.length} {t('个生效中')}
                  </Tag>
                ) : (
                  <Tag color='white' size='small' shape='circle'>
                    {t('无生效')}
                  </Tag>
                )}
                {allSubscriptions.length > activeSubscriptions.length && (
                  <Tag color='white' size='small' shape='circle'>
                    {allSubscriptions.length - activeSubscriptions.length}{' '}
                    {t('个已过期')}
                  </Tag>
                )}
              </div>
              <div className='flex items-center gap-2'>
                <Select
                  value={displayBillingPreference}
                  onChange={onChangeBillingPreference}
                  size='small'
                  optionList={[
                    {
                      value: 'subscription_first',
                      label: disableSubscriptionPreference
                        ? `${t('优先订阅')} (${t('无生效')})`
                        : t('优先订阅'),
                      disabled: disableSubscriptionPreference,
                    },
                    { value: 'wallet_first', label: t('优先钱包') },
                    {
                      value: 'subscription_only',
                      label: disableSubscriptionPreference
                        ? `${t('仅用订阅')} (${t('无生效')})`
                        : t('仅用订阅'),
                      disabled: disableSubscriptionPreference,
                    },
                    { value: 'wallet_only', label: t('仅用钱包') },
                  ]}
                />
                <Button
                  size='small'
                  theme='light'
                  type='tertiary'
                  icon={
                    <RefreshCw
                      size={12}
                      className={refreshing ? 'animate-spin' : ''}
                    />
                  }
                  onClick={handleRefresh}
                  loading={refreshing}
                />
              </div>
            </div>
            {disableSubscriptionPreference && isSubscriptionPreference && (
              <Text type='tertiary' size='small'>
                {t('已保存偏好为')}
                {subscriptionPreferenceLabel}
                {t('，当前无生效订阅，将自动使用钱包')}
              </Text>
            )}

            {hasAnySubscription ? (
              <>
                <Divider margin={8} />
                <div className='max-h-64 overflow-y-auto pr-1 semi-table-body'>
                  {allSubscriptions.map((sub, subIndex) => {
                    const isLast = subIndex === allSubscriptions.length - 1;
                    const subscription = sub.subscription;
                    const totalAmount = Number(subscription?.amount_total || 0);
                    const usedAmount = Number(subscription?.amount_used || 0);
                    const remainAmount =
                      totalAmount > 0
                        ? Math.max(0, totalAmount - usedAmount)
                        : 0;
                    const planTitle =
                      planTitleMap.get(subscription?.plan_id) || '';
                    const remainDays = getRemainingDays(sub);
                    const usagePercent = getUsagePercent(sub);
                    const now = Date.now() / 1000;
                    const isExpired = (subscription?.end_time || 0) < now;
                    const isCancelled = subscription?.status === 'cancelled';
                    const isActive =
                      subscription?.status === 'active' && !isExpired;

                    return (
                      <div key={subscription?.id || subIndex}>
                        {/* 订阅概要 */}
                        <div className='flex items-center justify-between text-xs mb-2'>
                          <div className='flex items-center gap-2'>
                            <span className='font-medium'>
                              {planTitle
                                ? `${planTitle} · ${t('订阅')} #${subscription?.id}`
                                : `${t('订阅')} #${subscription?.id}`}
                            </span>
                            {isActive ? (
                              <Tag
                                color='white'
                                size='small'
                                shape='circle'
                                prefixIcon={<Badge dot type='success' />}
                              >
                                {t('生效')}
                              </Tag>
                            ) : isCancelled ? (
                              <Tag color='white' size='small' shape='circle'>
                                {t('已作废')}
                              </Tag>
                            ) : (
                              <Tag color='white' size='small' shape='circle'>
                                {t('已过期')}
                              </Tag>
                            )}
                          </div>
                          {isActive && (
                            <span className='text-gray-500'>
                              {t('剩余')} {remainDays} {t('天')}
                            </span>
                          )}
                        </div>
                        <div className='text-xs text-gray-500 mb-2'>
                          {isActive
                            ? t('至')
                            : isCancelled
                              ? t('作废于')
                              : t('过期于')}{' '}
                          {new Date(
                            (subscription?.end_time || 0) * 1000,
                          ).toLocaleString()}
                        </div>
                        {isActive && subscription?.next_reset_time > 0 && (
                          <div className='text-xs text-gray-500 mb-2'>
                            {t('下一次重置')}:{' '}
                            {new Date(
                              subscription.next_reset_time * 1000,
                            ).toLocaleString()}
                          </div>
                        )}
                        <div className='text-xs text-gray-500 mb-2'>
                          {t('总额度')}:{' '}
                          {totalAmount > 0 ? (
                            <Tooltip
                              content={`${t('原生额度')}：${usedAmount}/${totalAmount} · ${t('剩余')} ${remainAmount}`}
                            >
                              <span>
                                {renderQuota(usedAmount)}/
                                {renderQuota(totalAmount)} · {t('剩余')}{' '}
                                {renderQuota(remainAmount)}
                              </span>
                            </Tooltip>
                          ) : (
                            t('不限')
                          )}
                          {totalAmount > 0 && (
                            <span className='ml-2'>
                              {t('已用')} {usagePercent}%
                            </span>
                          )}
                        </div>
                        {!isLast && <Divider margin={12} />}
                      </div>
                    );
                  })}
                </div>
              </>
            ) : (
              <div className='text-xs text-gray-500'>
                {t('购买套餐后即可享受模型权益')}
              </div>
            )}
          </Card>

          {/* 可购买套餐 - 标准定价卡片 */}
          {plans.length > 0 ? (
            <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-5 w-full px-1'>
              {plans.map((p, index) => {
                const plan = p?.plan;
                const totalAmount = Number(plan?.total_amount || 0);
                const displayPrice = formatSubscriptionPrice(plan);
                const isPopular = index === 0 && plans.length > 1;
                const limit = Number(plan?.max_purchase_per_user || 0);
                const availability = planSaleAvailability(p);
                const canPurchase = planCanBePurchased(p);
                const unavailableText = planUnavailableText(p, t);
                const weeklyDaysLabel = weeklySaleDaysLabel(
                  availability?.weekly_sale_days,
                  t,
                );
                const countdownText = formatCountdown(
                  availability?.daily_sale_countdown_seconds,
                  t,
                );
                const limitLabel = limit > 0 ? `${t('限购')} ${limit}` : null;
                const dailyWindowLabel =
                  plan?.daily_sale_starts_at && plan?.daily_sale_ends_at
                    ? `${t('每日售卖')}: ${plan.daily_sale_starts_at}-${plan.daily_sale_ends_at}`
                    : null;
                const weeklyLabel = weeklyDaysLabel
                  ? `${t('每周上架')}: ${weeklyDaysLabel}`
                  : null;
                const totalLabel =
                  totalAmount > 0
                    ? `${t('总额度')}: ${renderQuota(totalAmount)}`
                    : `${t('总额度')}: ${t('不限')}`;
                const upgradeLabel = plan?.upgrade_group
                  ? `${t('升级分组')}: ${plan.upgrade_group}`
                  : null;
                const resetLabel =
                  formatSubscriptionResetPeriod(plan, t) === t('不重置')
                    ? null
                    : `${t('额度重置')}: ${formatSubscriptionResetPeriod(plan, t)}`;
                const planBenefits = [
                  {
                    label: `${t('有效期')}: ${formatSubscriptionDuration(plan, t)}`,
                  },
                  resetLabel ? { label: resetLabel } : null,
                  totalAmount > 0
                    ? {
                        label: totalLabel,
                        tooltip: `${t('原生额度')}：${totalAmount}`,
                      }
                    : { label: totalLabel },
                  limitLabel ? { label: limitLabel } : null,
                  dailyWindowLabel ? { label: dailyWindowLabel } : null,
                  weeklyLabel ? { label: weeklyLabel } : null,
                  !canPurchase && unavailableText
                    ? {
                        label: countdownText
                          ? `${unavailableText} · ${t('剩余')} ${countdownText}`
                          : unavailableText,
                      }
                    : null,
                  upgradeLabel ? { label: upgradeLabel } : null,
                ].filter(Boolean);

                return (
                  <Card
                    key={plan?.id}
                    className={`!rounded-xl transition-all hover:shadow-lg w-full h-full ${
                      isPopular ? 'ring-2 ring-purple-500' : ''
                    }`}
                    bodyStyle={{ padding: 0 }}
                  >
                    <div className='p-4 h-full flex flex-col'>
                      {/* 推荐标签 */}
                      {isPopular && (
                        <div className='mb-2'>
                          <Tag color='purple' shape='circle' size='small'>
                            <Sparkles size={10} className='mr-1' />
                            {t('推荐')}
                          </Tag>
                        </div>
                      )}
                      {/* 套餐名称 */}
                      <div className='mb-3'>
                        <Typography.Title
                          heading={5}
                          ellipsis={{ rows: 1, showTooltip: true }}
                          style={{ margin: 0 }}
                        >
                          {plan?.title || t('订阅套餐')}
                        </Typography.Title>
                        {plan?.subtitle && (
                          <Text
                            type='tertiary'
                            size='small'
                            ellipsis={{ rows: 1, showTooltip: true }}
                            style={{ display: 'block' }}
                          >
                            {plan.subtitle}
                          </Text>
                        )}
                      </div>

                      {/* 价格区域 */}
                      <div className='py-2'>
                        <div className='flex items-baseline justify-start'>
                          <span className='text-3xl font-bold text-purple-600'>
                            {displayPrice}
                          </span>
                        </div>
                      </div>

                      {/* 套餐权益描述 */}
                      <div className='flex flex-col items-start gap-1 pb-2'>
                        {planBenefits.map((item) => {
                          const content = (
                            <div className='flex items-center gap-2 text-xs text-gray-500'>
                              <Badge dot type='tertiary' />
                              <span>{item.label}</span>
                            </div>
                          );
                          if (!item.tooltip) {
                            return (
                              <div
                                key={item.label}
                                className='w-full flex justify-start'
                              >
                                {content}
                              </div>
                            );
                          }
                          return (
                            <Tooltip key={item.label} content={item.tooltip}>
                              <div className='w-full flex justify-start'>
                                {content}
                              </div>
                            </Tooltip>
                          );
                        })}
                      </div>

                      <div className='mt-auto'>
                        <Divider margin={12} />

                        {/* 购买按钮 */}
                        {(() => {
                          const count = getPlanPurchaseCount(p?.plan?.id);
                          const reached = limit > 0 && count >= limit;
                          const blocked = reached || !canPurchase;
                          const tip = reached
                            ? t('已达到购买上限') + ` (${count}/${limit})`
                            : unavailableText;
                          const label = reached
                            ? t('已达上限')
                            : !canPurchase
                              ? t('暂不可购买')
                              : t('立即订阅');
                          const buttonEl = (
                            <Button
                              theme='outline'
                              type='primary'
                              block
                              disabled={blocked}
                              onClick={() => {
                                if (!blocked) openBuy(p);
                              }}
                            >
                              {label}
                            </Button>
                          );
                          return blocked ? (
                            <Tooltip content={tip} position='top'>
                              {buttonEl}
                            </Tooltip>
                          ) : (
                            buttonEl
                          );
                        })()}
                      </div>
                    </div>
                  </Card>
                );
              })}
            </div>
          ) : (
            <div className='text-center text-gray-400 text-sm py-4'>
              {t('暂无可购买套餐')}
            </div>
          )}
        </Space>
      )}
    </>
  );

  return (
    <>
      {withCard ? (
        <Card className='!rounded-2xl shadow-sm border-0'>{cardContent}</Card>
      ) : (
        <div className='space-y-3'>{cardContent}</div>
      )}

      {/* 购买确认弹窗 */}
      <SubscriptionPurchaseModal
        t={t}
        visible={open}
        onCancel={closeBuy}
        selectedPlan={selectedPlan}
        paying={paying}
        selectedEpayMethod={selectedEpayMethod}
        setSelectedEpayMethod={setSelectedEpayMethod}
        epayMethods={epayMethods}
        userQuota={userQuota}
        enableOnlineTopUp={enableOnlineTopUp}
        enableStripeTopUp={enableStripeTopUp}
        enableCreemTopUp={enableCreemTopUp}
        enableWechatPayTopUp={enableWechatPayTopUp}
        enableAlipayTopUp={enableAlipayTopUp}
        enableCasdoorTopUp={enableCasdoorTopUp}
        officialPaySupported={planSupportsOfficialPay(selectedPlan?.plan)}
        purchaseLimitInfo={
          selectedPlan?.plan?.id
            ? {
                limit: Number(selectedPlan?.plan?.max_purchase_per_user || 0),
                count: getPlanPurchaseCount(selectedPlan?.plan?.id),
              }
            : null
        }
        saleAvailability={planSaleAvailability(selectedPlan)}
        saleBlockText={planUnavailableText(selectedPlan, t)}
        onPayBalance={payBalance}
        onPayStripe={payStripe}
        onPayCreem={payCreem}
        onPayEpay={payEpay}
        onPayWechatPay={() => payOfficial('wechat_pay')}
        onPayAlipay={() => payOfficial('alipay_direct')}
        onPayCasdoor={() => payOfficial('casdoor')}
      />
      <OfficialPaymentModal
        t={t}
        visible={officialPaymentOpen}
        onCancel={() => setOfficialPaymentOpen(false)}
        payment={officialPayment}
      />
    </>
  );
};

export default SubscriptionPlansCard;
