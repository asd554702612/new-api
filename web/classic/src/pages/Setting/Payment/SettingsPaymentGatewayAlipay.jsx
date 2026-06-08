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

export default function SettingsPaymentGatewayAlipay(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle ? undefined : t('支付宝设置');
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    AlipayEnabled: false,
    AlipayUnitPrice: 0,
    AlipayAppID: '',
    AlipayPrivateKey: '',
    AlipayPublicKey: '',
    AlipaySandbox: false,
    AlipayNotifyURL: '',
    AlipayReturnURL: '',
    AlipayPageEnabled: true,
    AlipayWapEnabled: true,
    AlipayFaceEnabled: true,
  });
  const formApiRef = useRef(null);
  const toBool = (value) => value === true || value === 'true';

  useEffect(() => {
    if (!props.options || !formApiRef.current) return;
    const currentInputs = {
      AlipayEnabled: !!props.options.AlipayEnabled,
      AlipayUnitPrice:
        props.options.AlipayUnitPrice !== undefined
          ? parseFloat(props.options.AlipayUnitPrice)
          : 0,
      AlipayAppID: props.options.AlipayAppID || '',
      AlipayPrivateKey: '',
      AlipayPublicKey: props.options.AlipayPublicKey || '',
      AlipaySandbox: toBool(props.options.AlipaySandbox),
      AlipayNotifyURL: props.options.AlipayNotifyURL || '',
      AlipayReturnURL: props.options.AlipayReturnURL || '',
      AlipayPageEnabled:
        props.options.AlipayPageEnabled !== undefined
          ? !!props.options.AlipayPageEnabled
          : true,
      AlipayWapEnabled:
        props.options.AlipayWapEnabled !== undefined
          ? !!props.options.AlipayWapEnabled
          : true,
      AlipayFaceEnabled:
        props.options.AlipayFaceEnabled !== undefined
          ? !!props.options.AlipayFaceEnabled
          : true,
    };
    setInputs(currentInputs);
    formApiRef.current.setValues(currentInputs);
  }, [props.options]);

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

  const submitAlipaySetting = async () => {
    setLoading(true);
    try {
      const options = [
        { key: 'AlipayEnabled', value: inputs.AlipayEnabled ? 'true' : 'false' },
        {
          key: 'AlipayUnitPrice',
          value:
            inputs.AlipayUnitPrice !== undefined &&
            inputs.AlipayUnitPrice !== null
              ? inputs.AlipayUnitPrice.toString()
              : '0',
        },
        { key: 'AlipayAppID', value: inputs.AlipayAppID || '' },
        { key: 'AlipayPublicKey', value: inputs.AlipayPublicKey || '' },
        { key: 'AlipaySandbox', value: inputs.AlipaySandbox ? 'true' : 'false' },
        {
          key: 'AlipayNotifyURL',
          value: removeTrailingSlash(inputs.AlipayNotifyURL || ''),
        },
        {
          key: 'AlipayReturnURL',
          value: removeTrailingSlash(inputs.AlipayReturnURL || ''),
        },
        {
          key: 'AlipayPageEnabled',
          value: inputs.AlipayPageEnabled ? 'true' : 'false',
        },
        {
          key: 'AlipayWapEnabled',
          value: inputs.AlipayWapEnabled ? 'true' : 'false',
        },
        {
          key: 'AlipayFaceEnabled',
          value: inputs.AlipayFaceEnabled ? 'true' : 'false',
        },
      ];

      if ((inputs.AlipayPrivateKey || '').trim()) {
        options.push({ key: 'AlipayPrivateKey', value: inputs.AlipayPrivateKey });
      }

      if (await updateOptions(options)) {
        showSuccess(t('更新成功'));
        props.refresh?.();
      }
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
        onValueChange={setInputs}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={sectionTitle}>
          <Banner
            type='info'
            icon={<BookOpen size={16} />}
            description={
              <>
                {t('支付宝回调地址')}：
                {props.options.ServerAddress
                  ? removeTrailingSlash(props.options.ServerAddress)
                  : t('网站地址')}
                /api/alipay/notify
                <br />
                Return URL：
                {props.options.ServerAddress
                  ? removeTrailingSlash(props.options.ServerAddress)
                  : t('网站地址')}
                /api/alipay/return
              </>
            }
            style={{ marginBottom: 16 }}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='AlipayEnabled'
                label={t('启用支付宝')}
                checkedText='｜'
                uncheckedText='〇'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='AlipaySandbox'
                label={t('支付宝沙箱模式')}
                checkedText='｜'
                uncheckedText='〇'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input field='AlipayAppID' label='AppID' />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='AlipayUnitPrice'
                precision={2}
                label={t('充值价格（x元/美金）')}
                placeholder={t('例如：7，就是7元/美金')}
                extraText={t('按 1 美元对应的站内价格填写')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='AlipayPublicKey'
                label={t('支付宝公钥')}
                autosize
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='AlipayPrivateKey'
                label={t('应用私钥')}
                placeholder={t('留空表示保持当前不变')}
                autosize
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayNotifyURL'
                label='Notify URL'
                placeholder='https://gateway.example.com/api/alipay/notify'
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayReturnURL'
                label='Return URL'
                placeholder='https://gateway.example.com/api/alipay/return'
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch field='AlipayPageEnabled' label={t('电脑网站支付')} />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch field='AlipayWapEnabled' label={t('手机网站支付')} />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch field='AlipayFaceEnabled' label={t('当面付扫码')} />
            </Col>
          </Row>
          <Button onClick={submitAlipaySetting} style={{ marginTop: 16 }}>
            {t('更新支付宝设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
