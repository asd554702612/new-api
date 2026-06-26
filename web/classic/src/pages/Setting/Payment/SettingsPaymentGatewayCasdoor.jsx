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

import React, { useEffect, useRef, useState } from 'react';
import { Banner, Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import { BookOpen } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
} from '../../../helpers';

const defaultInputs = {
  CasdoorPaymentEnabled: false,
  CasdoorBaseURL: 'https://login.gepinkeji.com',
  CasdoorClientID: '',
  CasdoorClientSecret: '',
  CasdoorApplicationName: '',
  CasdoorPaymentProduct: 'external-pay-template',
  CasdoorPaymentProvider: 'provider_payment_wechat_gepinkeji',
  CasdoorPaymentCurrency: 'CNY',
  CasdoorPaymentUnitPrice: 0,
  CasdoorPaymentMinTopUp: 1,
};

export default function SettingsPaymentGatewayCasdoor(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle
    ? undefined
    : t('Casdoor 统一支付设置');
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState(defaultInputs);
  const formApiRef = useRef(null);

  useEffect(() => {
    if (!props.options || !formApiRef.current) return;
    const currentInputs = {
      CasdoorPaymentEnabled: !!props.options.CasdoorPaymentEnabled,
      CasdoorBaseURL:
        props.options.CasdoorBaseURL || defaultInputs.CasdoorBaseURL,
      CasdoorClientID: props.options.CasdoorClientID || '',
      CasdoorClientSecret: '',
      CasdoorApplicationName: props.options.CasdoorApplicationName || '',
      CasdoorPaymentProduct:
        props.options.CasdoorPaymentProduct ||
        defaultInputs.CasdoorPaymentProduct,
      CasdoorPaymentProvider:
        props.options.CasdoorPaymentProvider ||
        defaultInputs.CasdoorPaymentProvider,
      CasdoorPaymentCurrency:
        props.options.CasdoorPaymentCurrency ||
        defaultInputs.CasdoorPaymentCurrency,
      CasdoorPaymentUnitPrice:
        props.options.CasdoorPaymentUnitPrice !== undefined
          ? parseFloat(props.options.CasdoorPaymentUnitPrice)
          : 0,
      CasdoorPaymentMinTopUp:
        props.options.CasdoorPaymentMinTopUp !== undefined
          ? parseInt(props.options.CasdoorPaymentMinTopUp)
          : 1,
    };
    setInputs(currentInputs);
    formApiRef.current.setValues(currentInputs);
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const updateOptions = async (options) => {
    const results = await Promise.all(
      options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        }),
      ),
    );
    const errors = results.filter((res) => !res.data.success);
    if (errors.length > 0) {
      errors.forEach((res) => showError(res.data.message));
      return false;
    }
    return true;
  };

  const submitCasdoorSetting = async () => {
    const values = {
      ...inputs,
      ...(formApiRef.current?.getValues?.() || {}),
    };
    setLoading(true);
    try {
      const options = [
        {
          key: 'CasdoorPaymentEnabled',
          value: values.CasdoorPaymentEnabled ? 'true' : 'false',
        },
        {
          key: 'CasdoorBaseURL',
          value: removeTrailingSlash(values.CasdoorBaseURL || ''),
        },
        { key: 'CasdoorClientID', value: values.CasdoorClientID || '' },
        {
          key: 'CasdoorApplicationName',
          value: values.CasdoorApplicationName || '',
        },
        {
          key: 'CasdoorPaymentProduct',
          value: values.CasdoorPaymentProduct || 'external-pay-template',
        },
        {
          key: 'CasdoorPaymentProvider',
          value:
            values.CasdoorPaymentProvider ||
            'provider_payment_wechat_gepinkeji',
        },
        {
          key: 'CasdoorPaymentCurrency',
          value: (values.CasdoorPaymentCurrency || 'CNY').toUpperCase(),
        },
        {
          key: 'CasdoorPaymentUnitPrice',
          value:
            values.CasdoorPaymentUnitPrice !== undefined &&
            values.CasdoorPaymentUnitPrice !== null
              ? values.CasdoorPaymentUnitPrice.toString()
              : '0',
        },
        {
          key: 'CasdoorPaymentMinTopUp',
          value:
            values.CasdoorPaymentMinTopUp !== undefined &&
            values.CasdoorPaymentMinTopUp !== null
              ? values.CasdoorPaymentMinTopUp.toString()
              : '1',
        },
      ];
      if ((values.CasdoorClientSecret || '').trim()) {
        options.push({
          key: 'CasdoorClientSecret',
          value: values.CasdoorClientSecret,
        });
      }
      const ok = await updateOptions(options);
      if (!ok) return;
      showSuccess(t('更新成功'));
      props.refresh?.();
    } catch (error) {
      showError(t('更新失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={sectionTitle}>
          <Banner
            type='info'
            icon={<BookOpen size={16} />}
            description={
              <>
                {t('Casdoor 统一支付用于创建微信支付订单，支付成功通过业务系统 Webhook 完成充值或订阅。')}
                <br />
                {t('回调地址')}：
                {props.options.ServerAddress
                  ? removeTrailingSlash(props.options.ServerAddress)
                  : t('网站地址')}
                /api/casdoor/payment/webhook
              </>
            }
            style={{ marginBottom: 12 }}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='CasdoorPaymentEnabled'
                size='default'
                checkedText='｜'
                uncheckedText='〇'
                label={t('启用 Casdoor 统一支付')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CasdoorBaseURL'
                label={t('Casdoor 地址')}
                placeholder='https://login.gepinkeji.com'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CasdoorApplicationName'
                label={t('Application 名称')}
                placeholder={t('例如：app-token-gptk')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CasdoorClientID'
                label='Client ID'
                placeholder={t('Casdoor Application clientId')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CasdoorClientSecret'
                label='Client Secret'
                type='password'
                placeholder={t('填写后覆盖当前密钥，留空表示保持当前不变')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='CasdoorPaymentUnitPrice'
                precision={2}
                label={t('充值价格（x元/美金）')}
                placeholder={t('留空或 0 时使用通用价格')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={6} lg={6} xl={6}>
              <Form.Input
                field='CasdoorPaymentProduct'
                label='Product'
                placeholder='external-pay-template'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CasdoorPaymentProvider'
                label='Provider'
                placeholder='provider_payment_wechat_gepinkeji'
              />
            </Col>
            <Col xs={24} sm={24} md={4} lg={4} xl={4}>
              <Form.Input field='CasdoorPaymentCurrency' label='Currency' />
            </Col>
            <Col xs={24} sm={24} md={6} lg={6} xl={6}>
              <Form.InputNumber
                field='CasdoorPaymentMinTopUp'
                label={t('最低充值美元数量')}
                placeholder='1'
              />
            </Col>
          </Row>
          <Button style={{ marginTop: 16 }} onClick={submitCasdoorSetting}>
            {t('更新 Casdoor 统一支付设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
