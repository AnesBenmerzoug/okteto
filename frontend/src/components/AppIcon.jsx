import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Favicon from 'react-favicon';

/* eslint-disable */
const icons = {
  default: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAABX1BMVEX///8AzMwAwboAwrgAwbkAwbkAwbkAwbgAwbkAwbkAwboAwrkAubkAvbYAwrkAwbkAwboAwbkAqqoAwrkAwbkAwbkAwbkAv78AxLoAwbkAwbkAw7wAwbkAwbkAwbkAwbkAwbkAwroAwLoAwbkAwbkAwbkAwbgAtrYAu7sAwbgAwLkAwrkAwLgAxrgA1aoAwbkAv7YAwbkAwrkA//8AxsYAwbkA/4AAwbkAwbkAwroAzLMAwbkAwbkAwrgAwbgAxbkAwbkAw7QAwrkAwbsAwbkAwbkAwboAwbkAwLgAwLgAw7wAwbkAwLkAv7gAwrcAwbkAwLkAwbgAwbkAwbkAwbsAv7cAwbkAwbgAwbkAyLYAwrgAwrkAwbkAwbkAv78AwbkAwbkAwbgAvLwAwbkAwbkAvbUAwbkAwbkAwLoAxLsAwbkAwbkAwLkAwbkAwbkAwbkAwLkAwbkAwLoAwbn///+n2QtZAAAAc3RSTlMABU6Nuubz6+PAiFALI5LvjCEDdfb7fwQay8gi3uTxwZFkcrLyxnMHD3eveZsSBqQc+ZYBCaUC9P5gCu7qGVYWfBG3Kbnak5VBZSbMjkQuz2qQmaAxIPWUew5LcezECHS92RPd6B+++nYemOeG0en42ItRa/lXggAAAAFiS0dEAIgFHUgAAAAJcEhZcwAADdcAAA3XAUIom3gAAAAHdElNRQfiCAIPFhV3OIViAAABfElEQVQ4y3VTZ1fCMBQNe7kABwoIFKQORAQUhaKIshRFRcW999b+/3MkL5GWNn1f8t699yRvBaGO6fQGo8lssdrsjh6ktt4+i9ix/gGngna5B8UuGxoekfOeUVFlY16Jd/pEhvk7z3j9FBoPBENcXzhCQ5+Hvj9BM4vyBJicMhFkmsQzJIrNSm/G5wAyx3GQmIcg2VVVKokxUxr7duAXFhWFuzPi0jJ2sjnMC2lV5/Irq3A64IIC0jR4IbKmLbBhQZFJrW+UyllkxYIKi+fx9GIIZhhlCarQb2TGR40l2MTMFoKuBpk51NvT2UZGLAiwC9jRt8dlgDHy2mXqoVGctkAnYEFO1alUzdLYBW8PrthX8k28mAfYPTwCRSvVJWgBeAw+RxamKRvoSZJgpxC56EoLBZoIf3ZOkItLAuRLdEsjxcpVqH6do+HN7f+NiTvW2t/LPsZDQ80/JuRJu7hMN/30rFzSl1dBot/elZ8XbqmWjR+fX9/hn1+dhP4B/2iqD62r9isAAAAldEVYdGRhdGU6Y3JlYXRlADIwMTgtMDgtMDJUMTU6MjI6MjErMDI6MDDf2b7yAAAAJXRFWHRkYXRlOm1vZGlmeQAyMDE4LTA4LTAyVDE1OjIyOjIxKzAyOjAwroQGTgAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAALdEVYdFRpdGxlAEdyb3Vwlh8ijAAAAFd6VFh0UmF3IHByb2ZpbGUgdHlwZSBpcHRjAAB4nOPyDAhxVigoyk/LzEnlUgADIwsuYwsTIxNLkxQDEyBEgDTDZAMjs1Qgy9jUyMTMxBzEB8uASKBKLgDqFxF08kI1lQAAAABJRU5ErkJggg=='
}; 
/* eslint-enable */

class AppIcon extends Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <Favicon url={icons[this.props.icon] || icons.default} />
    );
  }
}

AppIcon.defaultProps = {
  icon: 'default'
};

AppIcon.propTypes = {
  icon: PropTypes.string
};

export default AppIcon;